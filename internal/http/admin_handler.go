package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"url-shortener/internal/auth"
)

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminHandler struct {
	adminUser     string
	adminPassword string
	jwtMgr        *auth.JWTManager
}

func NewAdminHandler(adminUser, adminPassword string, jwtMgr *auth.JWTManager) *AdminHandler {
	return &AdminHandler{
		adminUser:     adminUser,
		adminPassword: adminPassword,
		jwtMgr:        jwtMgr,
	}
}


func (h *AdminHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.Username != h.adminUser || req.Password != h.adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.jwtMgr.GenerateAdminToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
	})
}
