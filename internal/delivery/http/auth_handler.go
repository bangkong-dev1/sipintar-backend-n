package http

import (
	"net/http"
	"strings"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/usecases"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	Usecase *usecases.AuthUseCase
	Logger  *zap.Logger
}

// Login godoc
// @Summary Login user
// @Description Login dengan email dan password, mengembalikan access token + refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body entities.LoginRequest true "Login payload"
// @Success 200 {object} entities.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req entities.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload tidak valid", "details": err.Error()})
		return
	}

	subdomain := strings.TrimSpace(c.GetHeader("X-subdomain"))
	resp, err := h.Usecase.Login(c.Request.Context(), req, subdomain)
	if err != nil {
		h.Logger.Warn("login failed", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Refresh godoc
// @Summary Refresh token
// @Description Mendapatkan access token baru menggunakan refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body entities.RefreshRequest true "Refresh token payload"
// @Success 200 {object} entities.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req entities.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload tidak valid", "details": err.Error()})
		return
	}

	resp, err := h.Usecase.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body entities.LogoutRequest true "Logout payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req entities.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload tidak valid", "details": err.Error()})
		return
	}

	if err := h.Usecase.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout berhasil"})
}
