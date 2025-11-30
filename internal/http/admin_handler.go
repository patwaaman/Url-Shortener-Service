package httpserver

import (
	"net/http"

	"url-shortener/internal/auth"
	"url-shortener/internal/dto"
	errconst "url-shortener/internal/error"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminHandler struct {
	adminUser     string
	adminPassword string
	jwtMgr        *auth.JWTManager
	log           *zap.Logger
}

func NewAdminHandler(adminUser, adminPassword string, jwtMgr *auth.JWTManager, log *zap.Logger) *AdminHandler {
	return &AdminHandler{
		adminUser:     adminUser,
		adminPassword: adminPassword,
		jwtMgr:        jwtMgr,
		log:           log,
	}
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errconst.ErrInvalidPayload.Error()})
		return
	}
	if req.Username != h.adminUser || req.Password != h.adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": errconst.ErrInvalidCredential.Error()})
		return
	}

	token, err := h.jwtMgr.GenerateAdminToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errconst.ErrTokenGeneration.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
	})
}
