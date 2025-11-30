package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	httpserver "url-shortener/internal/http"
)

// helper: creates a test router with rate limit middleware
func setupRouter(rps float64, burst int, t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	rl := httpserver.NewRateLimiter(rps, burst)
	t.Logf("Created RateLimiter with RPS=%v, Burst=%v", rps, burst)

	r := gin.New()
	r.Use(httpserver.RateLimitMiddleware(rl))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	return r
}

func performRequest(r *gin.Engine, t *testing.T, label string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	r.ServeHTTP(w, req)

	t.Logf("%s → response %d", label, w.Code)
	return w
}

//
// ───────────────────────────────────────────────
//  TEST 1: Burst capacity is honored
// ───────────────────────────────────────────────
//

func TestRateLimit_BurstAllowed(t *testing.T) {
	rps := float64(1)
	burst := 2

	r := setupRouter(rps, burst, t)

	t.Log("Starting Burst Test")

	// #1
	w1 := performRequest(r, t, "Request 1")
	if w1.Code != 200 {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	// #2
	w2 := performRequest(r, t, "Request 2")
	if w2.Code != 200 {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	// #3
	w3 := performRequest(r, t, "Request 3 (should be 429)")
	if w3.Code != 429 {
		t.Fatalf("expected 429, got %d", w3.Code)
	}
}

//
// ───────────────────────────────────────────────
//  TEST 2: Rate limit recovers after token refill
// ───────────────────────────────────────────────
//

func TestRateLimit_RefillAllowsRequest(t *testing.T) {
	rps := float64(1)
	burst := 1

	r := setupRouter(rps, burst, t)

	t.Log("Starting Refill Test")

	// #1 allowed
	w1 := performRequest(r, t, "Request 1")
	if w1.Code != 200 {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	// #2 blocked
	w2 := performRequest(r, t, "Request 2 (should be 429)")
	if w2.Code != 429 {
		t.Fatalf("expected 429, got %d", w2.Code)
	}

	// Sleep to refill
	t.Log("Sleeping 1.1 seconds for token refill...")
	time.Sleep(1100 * time.Millisecond)

	// #3 allowed
	w3 := performRequest(r, t, "Request 3 after refill")
	if w3.Code != 200 {
		t.Fatalf("expected 200 after refill, got %d", w3.Code)
	}
}

//
// ───────────────────────────────────────────────
//  TEST 3: Per-IP limiting — different IPs get their own buckets
// ───────────────────────────────────────────────
//

func TestRateLimit_PerIP(t *testing.T) {
	rps := float64(1)
	burst := 1

	gin.SetMode(gin.TestMode)
	rl := httpserver.NewRateLimiter(rps, burst)

	r := gin.New()
	r.Use(httpserver.RateLimitMiddleware(rl))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	t.Log("Starting Per-IP Test")

	// Helper: perform request with custom IP
	doReq := func(ip string, label string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", ip)
		r.ServeHTTP(w, req)

		t.Logf("%s from IP=%s → response %d", label, ip, w.Code)
		return w
	}

	// First request from 1.1.1.1
	w1 := doReq("1.1.1.1", "Request 1")
	if w1.Code != 200 {
		t.Fatalf("expected 200 for 1.1.1.1, got %d", w1.Code)
	}

	// First request from 2.2.2.2
	w2 := doReq("2.2.2.2", "Request 1")
	if w2.Code != 200 {
		t.Fatalf("expected 200 for 2.2.2.2, got %d", w2.Code)
	}

	// Second request from 1.1.1.1
	w3 := doReq("1.1.1.1", "Request 2 (should be 429)")
	if w3.Code != 429 {
		t.Fatalf("expected 429 for 1.1.1.1, got %d", w3.Code)
	}

	// Second request from 2.2.2.2
	w4 := doReq("2.2.2.2", "Request 2 (should be 429)")
	if w4.Code != 429 {
		t.Fatalf("expected 429 for 2.2.2.2, got %d", w4.Code)
	}
}
