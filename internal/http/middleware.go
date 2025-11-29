package httpserver

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/url-shortener/internal/auth"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        rate.Limit(rps),
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if l, ok := rl.limiters[ip]; ok {
		return l
	}
	lim := rate.NewLimiter(rl.r, rl.burst)
	rl.limiters[ip] = lim
	return lim
}

func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c.Request)
		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		c.Next()
		status := c.Writer.Status()
		log.Printf("%s %s -> %d (%v)", method, path, status, time.Since(start))
	}
}

func AdminAuthMiddleware(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")
		claims, err := jwtMgr.VerifyAdminToken(tokenStr)
		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				log.Printf("jwt validation error: %v", ve)
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("admin_username", claims.Username)
		c.Next()
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
