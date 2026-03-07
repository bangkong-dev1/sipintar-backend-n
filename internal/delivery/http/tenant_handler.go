package http

import (
	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/usecases"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TenantHandler struct {
	usecase *usecases.TenantUsecase
	logger  *zap.Logger
}

// NewTenantHandler = buat handler baru
func NewTenantHandler(usecase *usecases.TenantUsecase, logger *zap.Logger) *TenantHandler {
	return &TenantHandler{
		usecase: usecase,
		logger:  logger.Named("handler"),
	}
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Subdomain string `json:"subdomain" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	tenant, err := h.usecase.CreateTenant(c.Request.Context(), req.Name, req.Subdomain)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, tenant)
}

func (h *TenantHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	tenant, err := h.usecase.GetTenant(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Tenant tidak ditemukan"})
		return
	}

	c.JSON(200, tenant)
}

func (h *TenantHandler) List(c *gin.Context) {
	tenants, total, err := h.usecase.ListTenants(c.Request.Context(), 1, 100)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"tenants": tenants, "total": total})
}

func (h *TenantHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.usecase.DeleteTenant(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Tenant berhasil dihapus"})
}

func (h *TenantHandler) Suspend(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.usecase.SuspendTenant(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Tenant berhasil disuspend"})
}

func (h *TenantHandler) Activate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.usecase.ActivateTenant(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Tenant berhasil diaktifkan"})
}

// ResolveTenant = middleware untuk resolve tenant dari subdomain
func (h *TenantHandler) ResolveTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host

		h.logger.Debug("Resolving tenant",
			zap.String("host", host),
			zap.String("path", c.Request.URL.Path),
		)

		tenant, err := h.usecase.ResolveFromHost(c.Request.Context(), host)
		if err != nil {
			h.logger.Warn("Tenant tidak ditemukan",
				zap.String("host", host),
				zap.Error(err),
			)
			c.JSON(404, gin.H{"error": "Tenant tidak ditemukan", "host": host})
			c.Abort()
			return
		}

		if tenant.Status != entities.StatusActive {
			c.JSON(403, gin.H{"error": "Tenant tidak aktif", "status": tenant.Status})
			c.Abort()
			return
		}

		c.Set("tenant", tenant)
		c.Set("tenant_id", tenant.ID.String())
		c.Set("tenant_context", &entities.Context{
			TenantID:  tenant.ID,
			Subdomain: tenant.Subdomain,
			Status:    tenant.Status,
		})

		h.logger.Debug("Tenant resolved",
			zap.String("tenant_id", tenant.ID.String()),
			zap.String("subdomain", tenant.Subdomain),
		)

		c.Next()
	}
}

// GetTenantFromGin = helper buat  ambil Tenant dari gin context
func GetTenantFromGin(c *gin.Context) *entities.Tenant {
	t, exists := c.Get("tenant")
	if !exists {
		return nil
	}
	return t.(*entities.Tenant)
}

// GetTenantContextFromGin = helper buat  ambil Context dari gin context
func GetTenantContextFromGin(c *gin.Context) *entities.Context {
	ctx, exists := c.Get("tenant_context")
	if !exists {
		return nil
	}
	return ctx.(*entities.Context)
}

// GetTenantIDFromGin = helper buat  ambil tenant ID dari gin context
func GetTenantIDFromGin(c *gin.Context) uuid.UUID {
	tenant := GetTenantFromGin(c)
	if tenant == nil {
		return uuid.Nil
	}
	return tenant.ID
}

// bere panic jika tenant tidak ada
func MustGetTenant(c *gin.Context) *entities.Tenant {
	tenant := GetTenantFromGin(c)
	if tenant == nil {
		panic("tenant tidak ditemukan di context")
	}
	return tenant
}
