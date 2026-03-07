package routes

import (
	"github.com/azril773/go-backend/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

func SetupTenantRoutes(r *gin.Engine, handler *http.TenantHandler) {
	admin := r.Group("/admin")
	{
		tenants := admin.Group("/tenants")
		{
			tenants.POST("", handler.Create)
			tenants.GET("", handler.List)
			tenants.GET("/:id", handler.GetByID)
			tenants.DELETE("/:id", handler.Delete)
			tenants.POST("/:id/suspend", handler.Suspend)
			tenants.POST("/:id/activate", handler.Activate)
		}
	}
	// ni buat example by domain coi seemlike tenant1.localhost:8080/api/ingfo
	api := r.Group("/api")
	api.Use(handler.ResolveTenant())
	{
		api.GET("/ingfo", func(c *gin.Context) {
			tenant := http.GetTenantFromGin(c)
			c.JSON(200, gin.H{
				"status": "ok",
				"tenant": gin.H{
					"id":        tenant.ID,
					"name":      tenant.Name,
					"subdomain": tenant.Subdomain,
				},
			})
		})
	}
}
