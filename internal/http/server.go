package httpserver

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/url-shortener/internal/analytics"
	"github.com/yourusername/url-shortener/internal/auth"
	"github.com/yourusername/url-shortener/internal/url"
)

func NewRouter(
	urlSvc url.Service,
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

	urlHandler := NewURLHandler(urlSvc, baseURL, analyticsSvc)
	adminHandler := NewAdminHandler(adminUser, adminPassword, jwtMgr)

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
