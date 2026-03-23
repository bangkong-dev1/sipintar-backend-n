package http

import (
	"net/http"
	"strings"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/usecases"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthMiddleware struct {
	Usecase *usecases.AuthUseCase
	Logger  *zap.Logger
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header tidak valid"})
			c.Abort()
			return
		}

		claims, err := m.Usecase.ValidateAccessToken(token)
		if err != nil {
			m.Logger.Warn("invalid access token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid atau kadaluarsa"})
			c.Abort()
			return
		}

		if paudCtx := GetPaudContextFromGin(c); paudCtx != nil {
			if claims.PaudID == nil || *claims.PaudID != paudCtx.PaudID {
				c.JSON(http.StatusForbidden, gin.H{"error": "token tidak sesuai tenant"})
				c.Abort()
				return
			}
		}

		c.Set("auth_claims", claims)
		c.Set("auth_user_id", claims.UserID.String())
		c.Set("auth_is_central", claims.IsCentral)
		if claims.PaudID != nil {
			c.Set("auth_paud_id", claims.PaudID.String())
		}
		if claims.RoleID != nil {
			c.Set("auth_role_id", claims.RoleID.String())
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireCentralAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetAuthClaimsFromGin(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "claim auth tidak ditemukan"})
			c.Abort()
			return
		}

		if !claims.IsCentral {
			c.JSON(http.StatusForbidden, gin.H{"error": "hanya akun central yang diizinkan"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireTenantAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetAuthClaimsFromGin(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "claim auth tidak ditemukan"})
			c.Abort()
			return
		}

		if claims.IsCentral {
			c.JSON(http.StatusForbidden, gin.H{"error": "akun central tidak dapat mengakses endpoint tenant"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func extractBearerToken(authHeader string) string {
	parts := strings.SplitN(strings.TrimSpace(authHeader), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func GetAuthClaimsFromGin(c *gin.Context) *entities.AccessTokenClaims {
	val, ok := c.Get("auth_claims")
	if !ok {
		return nil
	}
	claims, ok := val.(*entities.AccessTokenClaims)
	if !ok {
		return nil
	}
	return claims
}
