package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	DB *sqlx.DB
}

func (r *RoleRepository) List(ctx context.Context) ([]*entities.Role, error) {
	var roles []*entities.Role
	query := `SELECT id, name, created_at, updated_at FROM roles WHERE deleted_at IS NULL ORDER BY created_at DESC`
	err := r.DB.SelectContext(ctx, &roles, query)
	if err != nil {
		return nil, fmt.Errorf("gagal list roles: %w", err)
	}
	return roles, nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles WHERE id = $1 AND deleted_at IS NULL`
	role := &entities.Role{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("gagal get role: %w", err)
	}
	return role, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *entities.Role) error {
	if role.ID == uuid.Nil {
		role.ID = uuid.New()
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	role.CreatedAt = now
	role.UpdatedAt = now

	query := `INSERT INTO roles (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.ExecContext(ctx, query, role.ID, role.Name, role.CreatedAt, role.UpdatedAt)
	if err != nil {
		return fmt.Errorf("gagal create role: %w", err)
	}
	return nil
}

func (r *RoleRepository) Update(ctx context.Context, role *entities.Role) error {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	role.UpdatedAt = now

	query := `UPDATE roles SET name = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
	result, err := r.DB.ExecContext(ctx, query, role.Name, role.UpdatedAt, role.ID)
	if err != nil {
		return fmt.Errorf("gagal update role: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("role tidak ditemukan: %s", role.ID)
	}
	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	result, err := r.DB.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("gagal delete role: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("role tidak ditemukan: %s", id)
	}
	return nil
}
