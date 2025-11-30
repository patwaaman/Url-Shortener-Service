package urlshortener

import (
	"context"
	"errors"
	"time"

	"url-shortener/internal/cache"
	"url-shortener/internal/model"
)

var ErrInvalidURL = errors.New("invalid url")

type Service interface {
	Shorten(ctx context.Context, original, customAlias string) (*model.URL, error)
	Resolve(ctx context.Context, code string) (*model.URL, error)
	List(ctx context.Context, page, pageSize int) ([]model.URL, int64, error)
}

type service struct {
	repo     Repository
	cache    *cache.RedisClient
	cacheTTL time.Duration
}

func NewService(repo Repository, cacheClient *cache.RedisClient) Service {
	return &service{
		repo:     repo,
		cache:    cacheClient,
		cacheTTL: 24 * time.Hour,
	}
}

func (s *service) Shorten(ctx context.Context, original, customAlias string) (*model.URL, error) {
	normalized, err := normalizeURL(original)
	if err != nil {
		return nil, ErrInvalidURL
	}

	// idempotent: same URL -> same record
	if existing, err := s.repo.FindByOriginalURL(ctx, normalized); err == nil {
		return existing, nil
	} else if err != ErrNotFound {
		return nil, err
	}

	var code string
	if customAlias != "" {
		code = sanitizeAlias(customAlias)
	} else {
		code = generateShortCode(normalized)
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
