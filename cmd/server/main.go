package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/url-shortener/internal/analytics"
	"github.com/yourusername/url-shortener/internal/auth"
	"github.com/yourusername/url-shortener/internal/cache"
	"github.com/yourusername/url-shortener/internal/config"
	"github.com/yourusername/url-shortener/internal/database"
	httpserver "github.com/yourusername/url-shortener/internal/http"
	grpcserver "github.com/yourusername/url-shortener/internal/grpc"
	"github.com/yourusername/url-shortener/internal/url"

	pb "github.com/yourusername/url-shortener/proto"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	db := database.NewPostgres(cfg.DatabaseURL)
	redisClient := cache.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer redisClient.Close()

	// AutoMigrate
	if err := db.AutoMigrate(&url.URL{}, &url.ClickStat{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	urlRepo := url.NewRepository(db)
	urlSvc := url.NewService(urlRepo, redisClient)
	analyticsSvc := analytics.NewService(db)
	jwtMgr := auth.NewJWTManager(cfg.AdminJWTSecret, 24*time.Hour)

	rl := httpserver.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	router := httpserver.NewRouter(urlSvc, cfg.BaseURL, analyticsSvc, jwtMgr, cfg.AdminUser, cfg.AdminPassword, rl)

	httpSrv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// gRPC server
	grpcSrv := grpc.NewServer()
	grpcHandler := grpcserver.NewServer(urlSvc, cfg.BaseURL)
	pb.RegisterURLShortenerServer(grpcSrv, grpcHandler)

	// Start HTTP
	go func() {
		log.Printf("HTTP server listening on :%s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// Start gRPC
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("failed to listen for gRPC: %v", err)
		}
		log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
	grpcSrv.GracefulStop()

	log.Println("servers stopped")
}

