package httpserver

import (
	"time"

	"url-shortener/internal/analytics"
	"url-shortener/internal/auth"
	"url-shortener/internal/logger"
	"url-shortener/internal/urlshortener/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	urlSvc service.Service,
	baseURL string,
	analyticsSvc analytics.Service,
	jwtMgr *auth.JWTManager,
	adminUser, adminPassword string,
	rl *RateLimiter,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(LoggingMiddleware())
	r.Use(RateLimitMiddleware(rl))

	urlHandler := NewURLHandler(urlSvc, baseURL, analyticsSvc, logger.Log.Named("url-handler"))
	adminHandler := NewAdminHandler(adminUser, adminPassword, jwtMgr, logger.Log.Named("admin-handler"))

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "time": time.Now().UTC()})
	})

	// Public REST
	api := r.Group("/api/v1")
	{
		api.POST("/urls", urlHandler.Shorten)
	}

	// Redirect
	r.GET("/:code", urlHandler.Redirect)

	// Admin
	admin := r.Group("/api/v1/admin")
	admin.POST("/login", adminHandler.Login)

	protected := r.Group("/api/v1/admin")
	protected.Use(AdminAuthMiddleware(jwtMgr))
	{
		protected.GET("/urls", urlHandler.List)
		protected.GET("/analytics/clicks", urlHandler.AnalyticsClicks)
	}

	return r
}
