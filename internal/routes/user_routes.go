package routes

import (
	"github.com/azril773/go-backend/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

func UserRoute(r *gin.Engine, handler *http.UserHandler, paudHandler *http.PaudHandler, authMiddleware *http.AuthMiddleware) {
	admin := r.Group("/admin/api")
	admin.Use(
		paudHandler.RequireCentralDomain(),
		authMiddleware.RequireAuth(),
		authMiddleware.RequireCentralAccess(),
	)
	{
		users := admin.Group("/users")
		{
			users.POST("", handler.Create)
			users.GET("", handler.List)
			users.GET("/:id", handler.GetByID)
			users.PUT("/:id", handler.Update)
			users.DELETE("/:id", handler.Delete)
		}

		pauds := admin.Group("/pauds")
		{
			pauds.GET("/:subdomain/users", handler.GetUsersBySubdomain)
		}
	}

	api := r.Group("/api")
	api.Use(
		paudHandler.ResolvePaud(),
		authMiddleware.RequireAuth(),
		authMiddleware.RequireTenantAccess(),
	)
	{
		users := api.Group("/users")
		{
			users.GET("", handler.ListByPaud)
			users.GET("/:id", handler.GetByIDForPaud)
			users.POST("", handler.CreateForPaud)
		}
	}
}
