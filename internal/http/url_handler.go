package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"url-shortener/internal/analytics"
	"url-shortener/internal/urlshortener"
)

type shortenReq struct {
	OriginalURL string `json:"original_url" binding:"required"`
	CustomAlias string `json:"custom_alias"`
}

type shortenResp struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLHandler struct {
	svc       urlshortener.Service
	baseURL   string
	analytics analytics.Service
}

func NewURLHandler(svc urlshortener.Service, baseURL string, analyticsSvc analytics.Service) *URLHandler {
	return &URLHandler{svc: svc, baseURL: baseURL, analytics: analyticsSvc}
}

func (h *URLHandler) Shorten(c *gin.Context) {
	var req shortenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	u, err := h.svc.Shorten(c.Request.Context(), req.OriginalURL, req.CustomAlias)
	if err != nil {
		if err == urlshortener.ErrInvalidURL {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to shorten"})
		return
	}

	resp := shortenResp{
		ShortCode:   u.ShortCode,
		ShortURL:    h.baseURL + "/" + u.ShortCode,
		OriginalURL: u.OriginalURL,
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	u, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		if err == urlshortener.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve"})
		return
	}
	c.Redirect(http.StatusFound, u.OriginalURL)
}

func (h *URLHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	urls, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        urls,
		"page":        page,
		"page_size":   pageSize,
		"total_items": total,
	})
}

func (h *URLHandler) AnalyticsClicks(c *gin.Context) {
	fromStr := c.DefaultQuery("from", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	toStr := c.DefaultQuery("to", time.Now().Format("2006-01-02"))

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date"})
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date"})
		return
	}

	stats, err := h.analytics.GetDailyClicks(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get analytics"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"from":  fromStr,
		"to":    toStr,
		"items": stats,
	})
}
