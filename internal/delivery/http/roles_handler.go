package http

import (
	"net/http"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/usecases"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RoleHandler struct {
	Usecase *usecases.RoleUseCase
	Logger  *zap.Logger
}

// List godoc
// @Summary List all roles
// @Description Get list of all roles
// @Tags roles
// @Produce json
// @Success 200 {array} entities.Role
// @Failure 500 {object} map[string]string
// @Router /admin/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	roles, err := h.Usecase.ListRoles(ctx)
	if err != nil {
		h.Logger.Error("Failed to list roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}
	if len(roles) == 0 {
		c.JSON(http.StatusOK, []entities.Role{})
		return
	}
	c.JSON(http.StatusOK, roles)
}

// GetByID godoc
// @Summary Get role by ID
// @Description Get a single role by its ID
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} entities.Role
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/roles/{id} [get]
func (h *RoleHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Logger.Error("Invalid UUID", zap.Error(err), zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}

	role, err := h.Usecase.GetRoleByID(ctx, id)
	if err != nil {
		h.Logger.Error("Failed to get role", zap.Error(err), zap.String("id", idStr))
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}
	c.JSON(http.StatusOK, role)
}

// CreateRoleInput represents input for creating a role
type CreateRoleInput struct {
	Name string `json:"name" binding:"required"`
}

// Create godoc
// @Summary Create a new role
// @Description Create a new role with the provided name
// @Tags roles
// @Accept json
// @Produce json
// @Param role body CreateRoleInput true "Role data"
// @Success 201 {object} entities.Role
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var input CreateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.Logger.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	role, err := h.Usecase.CreateRole(ctx, input.Name)
	if err != nil {
		h.Logger.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}
	c.JSON(http.StatusCreated, role)
}

// UpdateRoleInput represents input for updating a role
type UpdateRoleInput struct {
	Name string `json:"name" binding:"required"`
}

// Update godoc
// @Summary Update a role
// @Description Update an existing role's name
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param role body UpdateRoleInput true "Role data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Logger.Error("Invalid UUID", zap.Error(err), zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}

	var input UpdateRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.Logger.Error("Invalid input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	err = h.Usecase.UpdateRole(ctx, id, input.Name)
	if err != nil {
		h.Logger.Error("Failed to update role", zap.Error(err), zap.String("id", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

// Delete godoc
// @Summary Delete a role
// @Description Delete a role by its ID
// @Tags roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Logger.Error("Invalid UUID", zap.Error(err), zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID format"})
		return
	}

	err = h.Usecase.DeleteRole(ctx, id)
	if err != nil {
		h.Logger.Error("Failed to delete role", zap.Error(err), zap.String("id", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}
