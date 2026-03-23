package routes

import (
	"github.com/azril773/go-backend/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

func AuthRoute(r *gin.Engine, handler *http.AuthHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.Refresh)
		auth.POST("/logout", handler.Logout)
	}
}
