package service

import (
	"context"
	"url-shortener/internal/model"
)

type Service interface {
	Shorten(ctx context.Context, original, customAlias string) (*model.URL, error)
	Resolve(ctx context.Context, code string) (*model.URL, error)
	List(ctx context.Context, page, pageSize int) ([]model.URL, int64, error)
}
