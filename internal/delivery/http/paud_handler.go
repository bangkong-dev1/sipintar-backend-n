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
	usecase *usecases.PaudUseCase
	logger  *zap.Logger
}

// NewPaudHandler = buat handler baru
func NewPaudHandler(usecase *usecases.PaudUseCase, logger *zap.Logger) *PaudHandler {
	return &PaudHandler{
		usecase: usecase,
		logger:  logger.Named("handler"),
	}
}

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
		h.logger.Warn("Logo tidak diupload", zap.Error(err))
		c.JSON(400, gin.H{"error": "Logo tidak diupload"})
		return
	}
	defer file.Close()
	ext := filepath.Ext(handler.Filename)
	subdomain := strings.ToLower(strings.ReplaceAll(req.Subdomain, " ", "-"))
	handler.Filename = fmt.Sprintf("%s%s", subdomain, ext)
	dst, err := os.Create(fmt.Sprintf("public/image/logo/%s", handler.Filename))
	if err != nil {
		h.logger.Warn("Gagal simpan logo", zap.Error(err))
		c.JSON(500, gin.H{"error": err.Error(), "filename": handler.Filename})
		return
	}
	defer dst.Close()

	paud, err := h.usecase.CreatePaud(c.Request.Context(), req.Name, req.Subdomain, fmt.Sprintf("/public/image/logo/%s", handler.Filename))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, paud)
}

func (h *PaudHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	paud, err := h.usecase.GetPaud(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Paud tidak ditemukan"})
		return
	}

	c.JSON(200, paud)
}

func (h *PaudHandler) List(c *gin.Context) {
	pauds, total, err := h.usecase.ListPaud(c.Request.Context(), 1, 100)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"pauds": pauds, "total": total})
}

func (h *PaudHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.usecase.DeletePaud(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Paud berhasil dihapus"})
}

func (h *PaudHandler) Suspend(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.usecase.SuspendPaud(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Paud berhasil disuspend"})
}

func (h *PaudHandler) Activate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.usecase.ActivatePaud(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Paud berhasil diaktifkan"})
}
func (h *PaudHandler) Test() string {
	return "test"
}

// ResolvePaud = middleware untuk resolve paud dari subdomain
func (h *PaudHandler) ResolvePaud() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		h.Test()
		h.logger.Debug("Resolving paud from host",
			zap.String("host", host),
			zap.String("path", c.Request.URL.Path),
		)

		paud, err := h.usecase.ResolveFromHeader(c.Request.Context(), c)
		if err != nil {
			h.logger.Warn("Paud tidak ditemukan",
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

		h.logger.Debug("Paud resolved",
			zap.String("paud_id", paud.ID.String()),
			zap.String("subdomain", paud.Subdomain),
		)

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
