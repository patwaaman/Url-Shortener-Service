package repository

import (
	"context"
	"time"
	"url-shortener/internal/model"
)

type Repository interface {
	FindByShortCode(ctx context.Context, code string) (*model.URL, error)
	FindByOriginalURL(ctx context.Context, original string) (*model.URL, error)
	Create(ctx context.Context, u *model.URL) error
	IncrementClick(ctx context.Context, id uint) error
	UpsertClickStat(ctx context.Context, urlID uint, day time.Time) error
	List(ctx context.Context, page, pageSize int) ([]model.URL, int64, error)
}
