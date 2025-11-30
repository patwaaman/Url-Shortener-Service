package database

import (
	"time"

	"url-shortener/internal/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func NewPostgres(dsn string) *gorm.DB {
	// Set GORM log level (warn by default)
	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
	}

	logger.Log.Info("initializing postgres connection", zap.String("dsn", dsn))

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		logger.Log.Fatal("failed to connect to postgres", zap.Error(err))
	}

	// Get underlying sql DB
	sqlDB, err := db.DB()
	if err != nil {
		logger.Log.Fatal("failed to get sql db from gorm", zap.Error(err))
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Log.Info("postgres connected successfully",
		zap.Int("max_open_conns", 25),
		zap.Int("max_idle_conns", 10),
		zap.Duration("conn_idle_time", 5*time.Minute),
		zap.Duration("conn_max_lifetime", time.Hour),
	)

	return db
}
