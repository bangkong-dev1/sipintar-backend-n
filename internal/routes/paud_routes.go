package routes

import (
	"github.com/azril773/go-backend/internal/delivery/http"
	"github.com/gin-gonic/gin"
)

func PaudRoute(r *gin.Engine, handler *http.PaudHandler) {
	admin := r.Group("/admin")
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
	// ni buat example by domain coi seemlike paud1.localhost:8080/api/ingfo
	api := r.Group("/api")
	api.Use(handler.ResolvePaud())
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
