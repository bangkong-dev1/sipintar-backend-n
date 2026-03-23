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

type UserRepository struct {
	DB *sqlx.DB
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `INSERT INTO users (id, paud_id, role_id, name, email, password, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.DB.ExecContext(ctx, query, user.ID, user.PaudID, user.RoleID, user.Name, user.Email, user.Password, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("gagal create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	query := `SELECT id, paud_id, role_id, name, email, password, created_at, updated_at 
			  FROM users WHERE id = $1`
	user := &entities.User{}
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.PaudID, &user.RoleID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user tidak ditemukan: %s", id)
		}
		return nil, fmt.Errorf("gagal get user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `SELECT id, paud_id, role_id, name, email, password, created_at, updated_at 
			  FROM users WHERE email = $1`
	user := &entities.User{}
	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.PaudID, &user.RoleID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user tidak ditemukan: %s", email)
		}
		return nil, fmt.Errorf("gagal get user by email: %w", err)
	}
	return user, nil
}

// GetByPaudID = get semua user berdasarkan paud_id (multi-tenant)
func (r *UserRepository) GetByPaudID(ctx context.Context, paudID uuid.UUID) ([]*entities.User, error) {
	query := `SELECT id, paud_id, role_id, name, email, password, created_at, updated_at 
			  FROM users WHERE paud_id = $1 ORDER BY created_at DESC`
	var users []*entities.User
	err := r.DB.SelectContext(ctx, &users, query, paudID)
	if err != nil {
		return nil, fmt.Errorf("gagal get users by paud_id: %w", err)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	user.UpdatedAt = time.Now().In(loc)

	query := `UPDATE users SET paud_id = $1, role_id = $2, name = $3, email = $4, updated_at = $5 WHERE id = $6`
	result, err := r.DB.ExecContext(ctx, query, user.PaudID, user.RoleID, user.Name, user.Email, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("gagal update user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user tidak ditemukan: %s", user.ID)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	query := `UPDATE users SET password = $1, updated_at = $2 WHERE id = $3`
	result, err := r.DB.ExecContext(ctx, query, hashedPassword, now, id)
	if err != nil {
		return fmt.Errorf("gagal update password: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user tidak ditemukan: %s", id)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal delete user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user tidak ditemukan: %s", id)
	}
	return nil
}

func (r *UserRepository) List(ctx context.Context, page, perPage int) ([]*entities.User, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	if err := r.DB.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count users: %w", err)
	}

	offset := (page - 1) * perPage
	query := `SELECT id, paud_id, role_id, name, email, password, created_at, updated_at 
			  FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	var users []*entities.User
	err := r.DB.SelectContext(ctx, &users, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal list users: %w", err)
	}

	return users, total, nil
}

func (r *UserRepository) ListByPaudID(ctx context.Context, paudID uuid.UUID, page, perPage int) ([]*entities.User, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE paud_id = $1`
	if err := r.DB.QueryRowContext(ctx, countQuery, paudID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal count users: %w", err)
	}

	offset := (page - 1) * perPage
	query := `SELECT id, paud_id, role_id, name, email, password, created_at, updated_at 
			  FROM users WHERE paud_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	var users []*entities.User
	err := r.DB.SelectContext(ctx, &users, query, paudID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal list users: %w", err)
	}

	return users, total, nil
}
