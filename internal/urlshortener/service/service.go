package service

import (
	"context"
	"time"

	"url-shortener/internal/cache"
	errconst "url-shortener/internal/error"
	"url-shortener/internal/model"
	"url-shortener/internal/urlshortener/helper"
	"url-shortener/internal/urlshortener/repository"

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
		// in real-world, inspect error for unique violation & maybe retry
		return nil, err
	}

	// warm cache
	if s.cache != nil {
		_ = s.cache.SetURL(ctx, u.ShortCode, u.OriginalURL, s.cacheTTL)
	}
	return u, nil
}

func (s *service) Resolve(ctx context.Context, code string) (*model.URL, error) {
	// 1. try redis
	if s.cache != nil {
		if orig, err := s.cache.GetURL(ctx, code); err == nil && orig != "" {
			// we still need metadata from DB if we want click_count, so we fetch
		}
	}

	u, err := s.repo.FindByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// increment click + stats
	_ = s.repo.IncrementClick(ctx, u.ID)
	_ = s.repo.UpsertClickStat(ctx, u.ID, time.Now().UTC())

	if s.cache != nil {
		_ = s.cache.SetURL(ctx, u.ShortCode, u.OriginalURL, s.cacheTTL)
	}
	return u, nil
}

func (s *service) List(ctx context.Context, page, pageSize int) ([]model.URL, int64, error) {
	return s.repo.List(ctx, page, pageSize)
}
