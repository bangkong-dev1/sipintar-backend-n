package http

import (
	"net/http"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/usecases"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserHandler struct {
	Usecase *usecases.UserUseCase
	Logger  *zap.Logger
}

// CreateUserInput = input untuk create user
type CreateUserInput struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	PaudID   *string `json:"paud_id,omitempty"`
	RoleID   *string `json:"role_id,omitempty"`
}

// Create godoc
// @Summary Create user baru
// @Description Membuat user baru
// @Tags Admin - User Management
// @Accept json
// @Produce json
// @Param user body CreateUserInput true "User data"
// @Success 201 {object} entities.User
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/api/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.Logger.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var paudID, roleID *uuid.UUID
	if input.PaudID != nil {
		parsed, err := uuid.Parse(*input.PaudID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid paud_id format"})
			return
		}
		paudID = &parsed
	}
	if input.RoleID != nil {
		parsed, err := uuid.Parse(*input.RoleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role_id format"})
			return
		}
		roleID = &parsed
	}

	user, err := h.Usecase.CreateUser(ctx, input.Name, input.Email, input.Password, paudID, roleID)
	if err != nil {
		h.Logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetByID godoc
// @Summary Get user by ID
// @Description Mendapatkan detail user berdasarkan ID
// @Tags Admin - User Management
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} entities.User
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	user, err := h.Usecase.GetUser(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// List godoc
// @Summary List semua user
// @Description Mendapatkan daftar semua user dengan pagination
// @Tags Admin - User Management
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/users [get]
func (h *UserHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	users, total, err := h.Usecase.ListUsers(ctx, 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "total": total})
}

// UpdateUserInput = input untuk update user
type UpdateUserInput struct {
	Name   string  `json:"name" binding:"required"`
	Email  string  `json:"email" binding:"required,email"`
	PaudID *string `json:"paud_id,omitempty"`
	RoleID *string `json:"role_id,omitempty"`
}

// Update godoc
// @Summary Update user
// @Description Update data user
// @Tags Admin - User Management
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param user body UpdateUserInput true "User data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.Logger.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	user, err := h.Usecase.GetUser(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	user.Name = input.Name
	user.Email = input.Email

	if input.PaudID != nil {
		parsed, err := uuid.Parse(*input.PaudID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid paud_id format"})
			return
		}
		user.PaudID = &parsed
	}
	if input.RoleID != nil {
		parsed, err := uuid.Parse(*input.RoleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role_id format"})
			return
		}
		user.RoleID = &parsed
	}

	if err := h.Usecase.UpdateUser(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil diupdate"})
}

// Delete godoc
// @Summary Delete user
// @Description Menghapus user berdasarkan ID
// @Tags Admin - User Management
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	if err := h.Usecase.DeleteUser(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil dihapus"})
}

// ListByPaud godoc
// @Summary List users by current paud
// @Description Mendapatkan daftar user untuk paud yang sedang diakses (dari subdomain)
// @Tags Tenant - User Management
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/users [get]
func (h *UserHandler) ListByPaud(c *gin.Context) {
	ctx := c.Request.Context()

	// Pakai helper dari paud untuk get paud context
	paud := GetPaudFromGin(c)
	if paud == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Paud context tidak ditemukan"})
		return
	}

	users, total, err := h.Usecase.ListUsersByPaudID(ctx, paud.ID, 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": total,
		"paud":  paud.Subdomain,
	})
}

// GetByIDForPaud godoc
// @Summary Get user by ID for current paud
// @Description Mendapatkan detail user yang termasuk dalam paud yang sedang diakses
// @Tags Tenant - User Management
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} entities.User
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/users/{id} [get]
func (h *UserHandler) GetByIDForPaud(c *gin.Context) {
	ctx := c.Request.Context()

	paud := GetPaudFromGin(c)
	if paud == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Paud context tidak ditemukan"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	user, err := h.Usecase.GetUser(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	// Pastikan user milik paud ini (multi-tenant isolation)
	if user.PaudID == nil || *user.PaudID != paud.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "User bukan bagian dari paud ini"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// nput untuk create user di tenant
type CreateUserForPaudInput struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	RoleID   *string `json:"role_id,omitempty"`
}

// CreateForPaud godoc
// @Summary Create user untuk current paud
// @Description Membuat user baru yang otomatis terikat dengan paud yang sedang diakses
// @Tags Tenant - User Management
// @Accept json
// @Produce json
// @Param user body CreateUserForPaudInput true "User data"
// @Success 201 {object} entities.User
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/users [post]
func (h *UserHandler) CreateForPaud(c *gin.Context) {
	ctx := c.Request.Context()

	paud := GetPaudFromGin(c)
	if paud == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Paud context tidak ditemukan"})
		return
	}

	var input CreateUserForPaudInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.Logger.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	var roleID *uuid.UUID
	if input.RoleID != nil {
		parsed, err := uuid.Parse(*input.RoleID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role_id format"})
			return
		}
		roleID = &parsed
	}

	// Otomatis set paud_id dari context
	user, err := h.Usecase.CreateUser(ctx, input.Name, input.Email, input.Password, &paud.ID, roleID)
	if err != nil {
		h.Logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUsersBySubdomain godoc
// @Summary Get users by paud subdomain
// @Description Mendapatkan daftar user berdasarkan subdomain paud
// @Tags Admin - User Management
// @Accept json
// @Produce json
// @Param subdomain path string true "Paud Subdomain"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/pauds/{subdomain}/users [get]
func (h *UserHandler) GetUsersBySubdomain(c *gin.Context) {
	ctx := c.Request.Context()
	subdomain := c.Param("subdomain")

	users, err := h.Usecase.GetUsersByPaudSubdomain(ctx, subdomain)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"users": []entities.User{}, "subdomain": subdomain})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":     users,
		"total":     len(users),
		"subdomain": subdomain,
	})
}
