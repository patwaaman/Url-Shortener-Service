package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"url-shortener/internal/analytics"
	"url-shortener/internal/dto"
	errconst "url-shortener/internal/error"
	"url-shortener/internal/urlshortener/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type URLHandler struct {
	svc       service.Service
	baseURL   string
	analytics analytics.Service
	log       *zap.Logger
}

const DATE_FORMAT = "2006-01-02"

func NewURLHandler(svc service.Service, baseURL string, analyticsSvc analytics.Service, log *zap.Logger) *URLHandler {
	return &URLHandler{svc: svc, baseURL: baseURL, analytics: analyticsSvc, log: log}
}

func (h *URLHandler) Shorten(c *gin.Context) {
	var req dto.ShortenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errconst.ErrInvalidPayload.Error()})
		return
	}

	u, err := h.svc.Shorten(c.Request.Context(), req.OriginalURL, req.CustomAlias)
	if err != nil {
		if err == errconst.ErrInvalidURL {
			c.JSON(http.StatusBadRequest, gin.H{"error": errconst.ErrInvalidURL.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errconst.ErrShortenUrlFailed.Error()})
		return
	}

	resp := dto.ShortenResp{
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
		if err == errconst.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": errconst.ErrShortCodeNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errconst.ErrResolveUrlFailed.Error()})
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

	fromStr := c.DefaultQuery("from", time.Now().AddDate(0, 0, -7).Format(DATE_FORMAT))
	toStr := c.DefaultQuery("to", time.Now().Format(DATE_FORMAT))

	from, err := time.Parse(DATE_FORMAT, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errconst.ErrInvalidFromDate.Error()})
		return
	}
	to, err := time.Parse(DATE_FORMAT, toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errconst.ErrInvalidToDate.Error()})
		return
	}

	stats, err := h.analytics.GetDailyClicks(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errconst.ErrAnalyticsStatsFailed.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"from":  fromStr,
		"to":    toStr,
		"items": stats,
	})
}
