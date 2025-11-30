package service

import (
	"context"
	"errors"
	"time"
	"url-shortener/internal/cache"
	errconst "url-shortener/internal/error"
	"url-shortener/internal/model"
	"url-shortener/internal/urlshortener/helper"
	"url-shortener/internal/urlshortener/repository"

	"github.com/jackc/pgconn"

	"go.uber.org/zap"
)

type service struct {
	repo     repository.Repository
	cache    *cache.RedisClient
	cacheTTL time.Duration
	log      *zap.Logger
}

func NewService(repo repository.Repository, cacheClient *cache.RedisClient, log *zap.Logger) Service {
	return &service{
		repo:     repo,
		cache:    cacheClient,
		cacheTTL: 24 * time.Hour,
		log:      log,
	}
}

func (s *service) Shorten(ctx context.Context, original, customAlias string) (*model.URL, error) {
	normalized, err := helper.NormalizeURL(original)
	if err != nil {
		return nil, errconst.ErrInvalidURL
	}

	// idempotent: same URL -> same record
	if existing, err := s.repo.FindByOriginalURL(ctx, normalized); err == nil {
		return existing, nil
	} else if err != errconst.ErrNotFound {
		return nil, err
	}

	var code string
	if customAlias != "" {
		code = helper.SanitizeAlias(customAlias)
	} else {
		code = helper.GenerateShortCode(normalized)
	}

	u := &model.URL{
		ShortCode:   code,
		OriginalURL: normalized,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			s.log.Error("enrty create error: ", zap.Error(pgErr))
			// 23505 => unique_violation
			if pgErr.Code == "23505" {
				return nil, errconst.ErrAliasTaken
			}
		}
		return nil, err
	}

	// update cache
	if s.cache != nil {
		if err := s.cache.SetURL(ctx, u.ShortCode, u.OriginalURL, s.cacheTTL); err != nil {
			s.log.Warn("failed to cache url", zap.Error(err))
		}
	}
	return u, nil
}

func (s *service) Resolve(ctx context.Context, code string) (*model.URL, error) {
	// 1. Try Redis fast path
	if s.cache != nil {
		if orig, err := s.cache.GetURL(ctx, code); err == nil && orig != "" {

			// DB stats update in background
			go func() {
				if dbURL, err := s.repo.FindByShortCode(context.Background(), code); err == nil {
					_ = s.repo.IncrementClick(context.Background(), dbURL.ID)
					_ = s.repo.UpsertClickStat(context.Background(), dbURL.ID, time.Now().UTC())
				}
			}()

			return &model.URL{
				ShortCode:   code,
				OriginalURL: orig,
			}, nil
		}
	}

	// 2. Redis MISS → fallback to DB
	u, err := s.repo.FindByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// update redis cache
	if s.cache != nil {
		if err := s.cache.SetURL(ctx, u.ShortCode, u.OriginalURL, s.cacheTTL); err != nil {
			s.log.Warn("failed to cache url:", zap.Error(err))
		}
	}

	// DB stats update in background
	go func(id uint) {
		_ = s.repo.IncrementClick(context.Background(), id)
		_ = s.repo.UpsertClickStat(context.Background(), id, time.Now().UTC())
	}(u.ID)

	return u, nil
}

func (s *service) List(ctx context.Context, page, pageSize int) ([]model.URL, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}
