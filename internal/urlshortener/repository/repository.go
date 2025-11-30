package repository

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
	errconst "url-shortener/internal/error"
	"url-shortener/internal/model"
)

type GormRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewRepository(db *gorm.DB, log *zap.Logger) Repository {
	return &GormRepository{db: db, log: log}
}

func (r *GormRepository) FindByShortCode(ctx context.Context, code string) (*model.URL, error) {
	var u model.URL
	if err := r.db.WithContext(ctx).Where("short_code = ?", code).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errconst.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) FindByOriginalURL(ctx context.Context, original string) (*model.URL, error) {
	var u model.URL
	if err := r.db.WithContext(ctx).Where("original_url = ?", original).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errconst.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *GormRepository) Create(ctx context.Context, u *model.URL) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *GormRepository) IncrementClick(ctx context.Context, id uint) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&model.URL{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"click_count":      gorm.Expr("click_count + 1"),
			"last_accessed_at": &now,
		}).Error
}

func (r *GormRepository) UpsertClickStat(ctx context.Context, urlID uint, day time.Time) error {
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stat model.ClickStat
		if err := tx.Where("url_id = ? AND date = ?", urlID, day).First(&stat).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				stat = model.ClickStat{
					URLID: urlID,
					Date:  day,
					Count: 1,
				}
				return tx.Create(&stat).Error
			}
			return err
		}
		stat.Count++
		return tx.Save(&stat).Error
	})
}

func (r *GormRepository) List(ctx context.Context, page, pageSize int) ([]model.URL, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var urls []model.URL
	var total int64

	tx := r.db.WithContext(ctx).Model(&model.URL{})
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&urls).Error; err != nil {
		return nil, 0, err
	}
	return urls, total, nil
}
