package routes

import (
	"github.com/azril773/go-backend/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

func PaudRoute(r *gin.Engine, handler *http.PaudHandler, authMiddleware *http.AuthMiddleware) {
	// Admin routes dengan middleware untuk memastikan hanya bisa diakses dari domain central
	admin := r.Group("/admin")
	admin.Use(
		handler.RequireCentralDomain(),
		authMiddleware.RequireAuth(),
		authMiddleware.RequireCentralAccess(),
	)
	{
		pauds := admin.Group("/pauds")
		{
			pauds.POST("", handler.Create)
			pauds.GET("", handler.List)
			pauds.GET("/:id", handler.GetByID)
			pauds.DELETE("/:id", handler.Delete)
			pauds.POST("/:id/suspend", handler.Suspend)
			pauds.POST("/:id/activate", handler.Activate)
		}
	}
	api := r.Group("/api")
	api.Use(
		handler.ResolvePaud(),
		authMiddleware.RequireAuth(),
		authMiddleware.RequireTenantAccess(),
	)
	{
		api.GET("/ingfo", func(c *gin.Context) {
			paud := http.GetPaudFromGin(c)
			c.JSON(200, gin.H{
				"status": "ok",
				"paud": gin.H{
					"id":        paud.ID,
					"name":      paud.Name,
					"logo":      paud.Logo,
					"subdomain": paud.Subdomain,
				},
			})
		})
	}
}
