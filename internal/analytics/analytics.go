package analytics

import (
	"context"
	"time"

	"url-shortener/internal/dto"
	"url-shortener/internal/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewService(db *gorm.DB, log *zap.Logger) Service {
	return &service{db: db, log: log}
}

func (s *service) GetDailyClicks(ctx context.Context, from, to time.Time) ([]dto.DailyClicks, error) {
	from = truncateDay(from)
	to = truncateDay(to)

	var rows []struct {
		Date  time.Time
		Count uint64
	}

	err := s.db.WithContext(ctx).
		Model(&model.ClickStat{}).
		Select("date, SUM(count) as count").
		Where("date BETWEEN ? AND ?", from, to).
		Group("date").
		Order("date").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	res := make([]dto.DailyClicks, len(rows))
	for i, r := range rows {
		res[i] = dto.DailyClicks{Date: r.Date, Count: r.Count}
	}
	return res, nil
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
