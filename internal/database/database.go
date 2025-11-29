package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres(dsn string) *gorm.DB {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql db: %v", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}
