package url

import (
	"time"

	"gorm.io/gorm"
)

type URL struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	ShortCode      string         `json:"short_code" gorm:"uniqueIndex;size:10;not null"`
	OriginalURL    string         `json:"original_url" gorm:"uniqueIndex;not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	ClickCount     uint64         `json:"click_count" gorm:"not null;default:0"`
	LastAccessedAt *time.Time     `json:"last_accessed_at,omitempty"`
}

type ClickStat struct {
	ID        uint      `gorm:"primaryKey"`
	URLID     uint      `gorm:"index;not null"`
	Date      time.Time `gorm:"index;not null"` // date only (truncate to day)
	Count     uint64    `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
