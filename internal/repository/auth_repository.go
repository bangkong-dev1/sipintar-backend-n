package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AuthRepository struct {
	DB *sqlx.DB
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token *entities.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	now := time.Now()
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	token.UpdatedAt = now

	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, jti, expires_at, revoked_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.DB.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.JTI,
		token.ExpiresAt,
		token.RevokedAt,
		token.CreatedAt,
		token.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("gagal menyimpan refresh token: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetActiveRefreshToken(ctx context.Context, tokenHash, jti string, userID uuid.UUID) (*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, jti, expires_at, revoked_at, created_at, updated_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND jti = $2 AND user_id = $3 AND revoked_at IS NULL AND expires_at > NOW()
		LIMIT 1
	`

	var refreshToken entities.RefreshToken
	if err := r.DB.GetContext(ctx, &refreshToken, query, tokenHash, jti, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("refresh token tidak valid")
		}
		return nil, fmt.Errorf("gagal mengambil refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (r *AuthRepository) RevokeRefreshTokenByID(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() WHERE id = $1 AND revoked_at IS NULL`
	_, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal revoke refresh token: %w", err)
	}
	return nil
}

func (r *AuthRepository) RevokeRefreshTokenByHash(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`
	_, err := r.DB.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("gagal revoke refresh token: %w", err)
	}
	return nil
}
