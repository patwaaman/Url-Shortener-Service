package analytics

import (
	"context"
	"time"

	"gorm.io/gorm"
	"url-shortener/internal/model"
)

type DailyClicks struct {
	Date  time.Time `json:"date"`
	Count uint64    `json:"count"`
}

type Service interface {
	GetDailyClicks(ctx context.Context, from, to time.Time) ([]DailyClicks, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) GetDailyClicks(ctx context.Context, from, to time.Time) ([]DailyClicks, error) {
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

	res := make([]DailyClicks, len(rows))
	for i, r := range rows {
		res[i] = DailyClicks{Date: r.Date, Count: r.Count}
	}
	return res, nil
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
