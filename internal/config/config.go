package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv      string
	HTTPPort    string
	BaseURL     string
	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	AdminJWTSecret string
	AdminUser      string
	AdminPassword  string

	RateLimitRPS   float64
	RateLimitBurst int

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func Load() *Config {
	return &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
		DatabaseURL:    mustEnv("DATABASE_URL"),
		RedisAddr:      getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		AdminJWTSecret: mustEnv("ADMIN_JWT_SECRET"),
		AdminUser:      mustEnv("ADMIN_USER"),
		AdminPassword:  mustEnv("ADMIN_PASSWORD"),
		RateLimitRPS:   getEnvFloat("RATE_LIMIT_RPS", 10),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 20),
		ReadTimeout:    getEnvDuration("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:   getEnvDuration("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:    getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env %s", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
