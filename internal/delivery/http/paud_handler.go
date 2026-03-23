package http

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/usecases"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PaudHandler struct {
	Usecase *usecases.PaudUseCase
	Logger  *zap.Logger
}

// Create godoc
// @Summary Create PAUD baru
// @Description Membuat PAUD baru dengan upload logo
// @Tags Admin - PAUD Management
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Nama PAUD"
// @Param subdomain formData string true "Subdomain PAUD"
// @Param Logo formData file true "Logo PAUD"
// @Success 201 {object} entities.Paud
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/pauds [post]
func (h *PaudHandler) Create(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Subdomain string `json:"subdomain" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Request.ParseMultipartForm(10 << 20)
	file, handler, err := c.Request.FormFile("Logo")
	if err != nil {
		h.Logger.Warn("Logo tidak diupload", zap.Error(err))
		c.JSON(400, gin.H{"error": "Logo tidak diupload"})
		return
	}
	defer file.Close()
	ext := filepath.Ext(handler.Filename)
	subdomain := strings.ToLower(strings.ReplaceAll(req.Subdomain, " ", "-"))
	handler.Filename = fmt.Sprintf("%s%s", subdomain, ext)
	dst, err := os.Create(fmt.Sprintf("public/image/logo/%s", handler.Filename))
	if err != nil {
		h.Logger.Warn("Gagal simpan logo", zap.Error(err))
		c.JSON(500, gin.H{"error": err.Error(), "filename": handler.Filename})
		return
	}
	defer dst.Close()

	paud, err := h.Usecase.CreatePaud(c.Request.Context(), req.Name, req.Subdomain, fmt.Sprintf("/public/image/logo/%s", handler.Filename))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, paud)
}

// GetByID godoc
// @Summary Get PAUD by ID
// @Description Mendapatkan detail PAUD berdasarkan ID
// @Tags Admin - PAUD Management
// @Accept json
// @Produce json
// @Param id path string true "PAUD ID (UUID)"
// @Success 200 {object} entities.Paud
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/pauds/{id} [get]
func (h *PaudHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	paud, err := h.Usecase.GetPaud(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Paud tidak ditemukan"})
		return
	}

	c.JSON(200, paud)
}

// List godoc
// @Summary List semua PAUD
// @Description Mendapatkan daftar semua PAUD dengan pagination
// @Tags Admin - PAUD Management
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/pauds [get]
func (h *PaudHandler) List(c *gin.Context) {
	pauds, total, err := h.Usecase.ListPaud(c.Request.Context(), 1, 100)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"pauds": pauds, "total": total})
}

// Delete godoc
// @Summary Delete PAUD
// @Description Menghapus PAUD berdasarkan ID
// @Tags Admin - PAUD Management
// @Accept json
// @Produce json
// @Param id path string true "PAUD ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/pauds/{id} [delete]
func (h *PaudHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.Usecase.DeletePaud(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Paud berhasil dihapus"})
}

// Suspend godoc
// @Summary Suspend PAUD
// @Description Menonaktifkan PAUD sementara
// @Tags Admin - PAUD Management
// @Accept json
// @Produce json
// @Param id path string true "PAUD ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/pauds/{id}/suspend [post]
func (h *PaudHandler) Suspend(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.Usecase.SuspendPaud(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Paud berhasil disuspend"})
}

// Activate godoc
// @Summary Activate PAUD
// @Description Mengaktifkan kembali PAUD yang di-suspend
// @Tags Admin - PAUD Management
// @Accept json
// @Produce json
// @Param id path string true "PAUD ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/pauds/{id}/activate [post]
func (h *PaudHandler) Activate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.Usecase.ActivatePaud(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Paud berhasil diaktifkan"})
}
func (h *PaudHandler) Test() string {
	return "test"
}

func (h *PaudHandler) RequireCentralDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host

		hostWithoutPort := strings.Split(host, ":")[0]

		h.Logger.Debug("Checking central domain access",
			zap.String("host", hostWithoutPort),
			zap.String("path", c.Request.URL.Path),
		)

		paud, err := h.Usecase.ResolveFromHeader(c.Request.Context(), c)
		if err == nil && paud != nil {
			h.Logger.Warn("Unauthorized access attempt from tenant subdomain",
				zap.String("subdomain", paud.Subdomain),
				zap.String("host", host),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(403, gin.H{
				"error":   "Akses ditolak: Endpoint admin hanya bisa diakses dari domain central",
				"message": "Access denied: Admin endpoints can only be accessed from central domain",
			})
			c.Abort()
			return
		}

		h.Logger.Debug("Central domain access granted",
			zap.String("host", hostWithoutPort),
		)

		c.Next()
	}
}

func (h *PaudHandler) ResolvePaud() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		h.Test()
		h.Logger.Debug("Resolving paud from host",
			zap.String("host", host),
			zap.String("path", c.Request.URL.Path),
		)

		paud, err := h.Usecase.ResolveFromHeader(c.Request.Context(), c)
		if err != nil {
			h.Logger.Warn("Paud tidak ditemukan",
				zap.String("host", host),
				zap.Error(err),
			)
			c.JSON(404, gin.H{"error": "Paud tidak ditemukan", "host": host})
			c.Abort()
			return
		}

		if paud.Status != entities.StatusActive {
			c.JSON(403, gin.H{"error": "Paud tidak aktif", "status": paud.Status})
			c.Abort()
			return
		}

		c.Set("paud", paud)
		c.Set("paud_id", paud.ID.String())
		c.Set("paud_context", &entities.Context{
			PaudID:    paud.ID,
			Subdomain: paud.Subdomain,
			Status:    paud.Status,
		})

		h.Logger.Debug("Paud resolved",
			zap.String("paud_id", paud.ID.String()),
			zap.String("subdomain", paud.Subdomain),
		)

		c.Next()

	}
}

// GetPaudFromGin = helper buat  ambil Paud dari gin context
func GetPaudFromGin(c *gin.Context) *entities.Paud {
	p, exists := c.Get("paud")
	if !exists {
		return nil
	}
	return p.(*entities.Paud)
}

// GetPaudContextFromGin = helper buat  ambil Context dari gin context
func GetPaudContextFromGin(c *gin.Context) *entities.Context {
	ctx, exists := c.Get("paud_context")
	if !exists {
		return nil
	}
	return ctx.(*entities.Context)
}

// GetPaudIDFromGin = helper buat  ambil paud ID dari gin context
func GetPaudIDFromGin(c *gin.Context) uuid.UUID {
	paud := GetPaudFromGin(c)
	if paud == nil {
		return uuid.Nil
	}
	return paud.ID
}

// bere panic jika paud tidak ada
func MustGetPaud(c *gin.Context) *entities.Paud {
	paud := GetPaudFromGin(c)
	if paud == nil {
		panic("paud tidak ditemukan di context")
	}
	return paud
}
