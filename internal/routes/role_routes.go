package routes

import (
	"github.com/azril773/go-backend/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

func RoleRoute(r *gin.Engine, handler *http.RoleHandler, paudHandler *http.PaudHandler, authMiddleware *http.AuthMiddleware) {
	admin := r.Group("/admin")
	admin.Use(
		paudHandler.RequireCentralDomain(),
		authMiddleware.RequireAuth(),
		authMiddleware.RequireCentralAccess(),
	)
	{
		roles := admin.Group("/roles")
		{
			roles.GET("", handler.List)
			roles.GET("/:id", handler.GetByID)
			roles.POST("", handler.Create)
			roles.PUT("/:id", handler.Update)
			roles.DELETE("/:id", handler.Delete)
		}
	}
}
