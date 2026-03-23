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

// PaudRepository = struct untuk akses database paud
type PaudRepository struct {
	DB *sqlx.DB
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

	query := `INSERT INTO pauds (id, name, subdomain, logo, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.ExecContext(ctx, query, t.ID, t.Name, t.Subdomain, t.Logo, t.Status, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("gagal create paud: %w", err)
	}
	return nil
}

func (r *PaudRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Paud, error) {
	query := `SELECT id, name, subdomain, logo, status, created_at, updated_at FROM pauds WHERE id = $1`
	t := &entities.Paud{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Name, &t.Subdomain, &t.Logo, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("paud tidak ditemukan: %s", id)
		}
		return nil, fmt.Errorf("gagal get paud: %w", err)
	}
	return t, nil
}

func (r *PaudRepository) GetBySubdomain(ctx context.Context, subdomain string) (*entities.Paud, error) {
	query := `SELECT id, name, subdomain, logo, status, created_at, updated_at 
	          FROM pauds WHERE subdomain = $1`

	t := &entities.Paud{}
	err := r.DB.GetContext(ctx, t, query, subdomain)
	if err != nil {
		return nil, fmt.Errorf("gagal get paud by subdomain: %w", err)
	}

	return t, nil
}

func (r *PaudRepository) Update(ctx context.Context, t *entities.Paud) error {
	t.UpdatedAt = time.Now()
	query := `UPDATE pauds SET name = $1, subdomain = $2, logo = $3, status = $4, updated_at = $5 WHERE id = $6`
	result, err := r.DB.ExecContext(ctx, query, t.Name, t.Subdomain, t.Logo, t.Status, t.UpdatedAt, t.ID)
	if err != nil {
		return fmt.Errorf("gagal update paud: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("paud tidak ditemukan: %s", t.ID)
	}
	return nil
}

func (r *PaudRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pauds WHERE id = $1`
	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal delete paud: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("paud tidak ditemukan: %s", id)
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
	countQuery := `SELECT COUNT(*) FROM pauds`
	if err := r.DB.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count pauds: %w", err)
	}

	offset := (page - 1) * perPage
	query := `SELECT id, name, subdomain, logo, status, created_at, updated_at FROM pauds ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.DB.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal list pauds: %w", err)
	}
	defer rows.Close()

	var pauds []*entities.Paud
	for rows.Next() {
		t := &entities.Paud{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Subdomain, &t.Logo, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan paud: %w", err)
		}
		pauds = append(pauds, t)
	}

	return pauds, total, nil
}
