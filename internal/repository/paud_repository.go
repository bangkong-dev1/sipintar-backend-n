package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PaudRepository = struct untuk akses database tenant
type PaudRepository struct {
	db *sqlx.DB
}

func NewPaudRepository(db *sqlx.DB) *PaudRepository {
	return &PaudRepository{db: db}
}

func (r *PaudRepository) Create(ctx context.Context, t *entities.Paud) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	if t.Status == "" {
		t.Status = entities.StatusPending
	}

	query := `INSERT INTO tenants (id, name, subdomain, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Subdomain, t.Status, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("gagal create tenant: %w", err)
	}
	return nil
}

func (r *PaudRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Paud, error) {
	query := `SELECT id, name, subdomain, status, created_at, updated_at FROM tenants WHERE id = $1`
	t := &entities.Paud{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Name, &t.Subdomain, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("tenant tidak ditemukan: %s", id)
		}
		return nil, fmt.Errorf("gagal get tenant: %w", err)
	}
	return t, nil
}

func (r *PaudRepository) GetBySubdomain(ctx context.Context, subdomain string) (*entities.Paud, error) {
	query := `SELECT id, name, subdomain, status, created_at, updated_at FROM tenants WHERE subdomain = $1`
	t := &entities.Paud{}
	err := r.db.QueryRowContext(ctx, query, subdomain).Scan(&t.ID, &t.Name, &t.Subdomain, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("tenant tidak ditemukan untuk subdomain: %s", subdomain)
		}
		return nil, fmt.Errorf("gagal get tenant by subdomain: %w", err)
	}
	return t, nil
}

func (r *PaudRepository) Update(ctx context.Context, t *entities.Paud) error {
	t.UpdatedAt = time.Now()
	query := `UPDATE tenants SET name = $1, subdomain = $2, status = $3, updated_at = $4 WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, t.Name, t.Subdomain, t.Status, t.UpdatedAt, t.ID)
	if err != nil {
		return fmt.Errorf("gagal update tenant: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tenant tidak ditemukan: %s", t.ID)
	}
	return nil
}

func (r *PaudRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal delete tenant: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tenant tidak ditemukan: %s", id)
	}
	return nil
}

func (r *PaudRepository) List(ctx context.Context, page, perPage int) ([]*entities.Paud, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM tenants`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count tenants: %w", err)
	}

	offset := (page - 1) * perPage
	query := `SELECT id, name, subdomain, status, created_at, updated_at FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*entities.Paud
	for rows.Next() {
		t := &entities.Paud{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Subdomain, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan tenant: %w", err)
		}
		tenants = append(tenants, t)
	}

	return tenants, total, nil
}
