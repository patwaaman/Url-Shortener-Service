package analytics

import (
	"context"
	"time"
	"url-shortener/internal/dto"
)

type Service interface {
	GetDailyClicks(ctx context.Context, from, to time.Time) ([]dto.DailyClicks, error)
}
