package usecases

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type PaudUseCase struct {
	DB     *sqlx.DB
	Repo   *repository.PaudRepository
	Logger *zap.Logger
}

func (u *PaudUseCase) CreatePaud(ctx context.Context, name, subdomain, logo string) (*entities.Paud, error) {
	if err := u.ValidateSubdomain(subdomain); err != nil {
		return nil, err
	}
	//  handle logo upload

	paud := &entities.Paud{
		ID:        uuid.New(),
		Name:      name,
		Subdomain: subdomain,
		Logo:      &logo,
		Status:    entities.StatusPending,
	}

	if err := u.Repo.Create(ctx, paud); err != nil {
		return nil, err
	}

	paud.Status = entities.StatusActive
	if err := u.Repo.Update(ctx, paud); err != nil {
		return nil, err
	}

	u.Logger.Info("paud created",
		zap.String("paud_id", paud.ID.String()),
		zap.String("name", name),
		zap.String("subdomain", subdomain),
	)

	return paud, nil
}

func (u *PaudUseCase) GetPaud(ctx context.Context, id uuid.UUID) (*entities.Paud, error) {
	return u.Repo.GetByID(ctx, id)
}

func (u *PaudUseCase) GetBySubdomain(ctx context.Context, subdomain string) (*entities.Paud, error) {
	return u.Repo.GetBySubdomain(ctx, subdomain)
}

func (u *PaudUseCase) UpdatePaud(ctx context.Context, paud *entities.Paud) error {
	if err := u.Repo.Update(ctx, paud); err != nil {
		return err
	}
	u.Logger.Info("paud updated", zap.String("paud_id", paud.ID.String()))
	return nil
}

func (u *PaudUseCase) DeletePaud(ctx context.Context, id uuid.UUID) error {
	if err := u.Repo.Delete(ctx, id); err != nil {
		return err
	}

	u.Logger.Info("paud deleted", zap.String("paud_id", id.String()))
	return nil
}

func (u *PaudUseCase) ListPaud(ctx context.Context, page, perPage int) ([]*entities.Paud, int, error) {
	return u.Repo.List(ctx, page, perPage)
}

func (u *PaudUseCase) SuspendPaud(ctx context.Context, id uuid.UUID) error {
	paud, err := u.Repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	paud.Status = entities.StatusSuspended
	return u.Repo.Update(ctx, paud)
}

func (u *PaudUseCase) ActivatePaud(ctx context.Context, id uuid.UUID) error {
	paud, err := u.Repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	paud.Status = entities.StatusActive
	return u.Repo.Update(ctx, paud)
}

func (u *PaudUseCase) GetPaudConn(ctx context.Context, paudID uuid.UUID) (*sql.Conn, error) {
	conn, err := u.DB.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal get connection: %w", err)
	}

	return conn, nil
}

// ResolveFromHost = extract paud dari host header
func (u *PaudUseCase) ResolveFromHost(ctx context.Context, host string) (*entities.Paud, error) {
	subdomain, err := u.ExtractSubdomain(host)
	if err != nil {
		return nil, err
	}

	paud, err := u.Repo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, fmt.Errorf("paud tidak ditemukan untuk subdomain: %s", subdomain)
	}

	return paud, nil
}

// Todo ResolveFromHeader
func (u *PaudUseCase) ResolveFromHeader(ctx context.Context, c *gin.Context) (*entities.Paud, error) {
	subdomain, err := u.ExtractFromHeader(c)
	if err != nil {
		return nil, err
	}

	paud, err := u.Repo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, fmt.Errorf("%s", err.Error())
	}

	return paud, nil
}

func (u *PaudUseCase) ExtractFromHeader(c *gin.Context) (string, error) {
	subdomain := c.GetHeader("X-subdomain")
	if subdomain == "" {
		return "", fmt.Errorf("no paud header found: %s", subdomain)
	}

	if err := u.ValidateSubdomain(subdomain); err != nil {
		return "", fmt.Errorf("invalid subdomain: %w", err)
	}

	return subdomain, nil
}

func (u *PaudUseCase) ExtractSubdomain(host string) (string, error) {
	if host == "" {
		return "", errors.New("host kosong")
	}

	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("format host tidak valid: %s", host)
	}

	subdomain := parts[0]
	if err := u.ValidateSubdomain(subdomain); err != nil {
		return "", err
	}

	return subdomain, nil
}

/// resolve by header

func (u *PaudUseCase) ValidateSubdomain(subdomain string) error {
	if len(subdomain) < 3 || len(subdomain) > 50 {
		return errors.New("subdomain harus 3-50 karakter")
	}

	validSubdomain := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !validSubdomain.MatchString(subdomain) {
		return errors.New("subdomain hanya boleh huruf kecil, angka, dan dash (-)")
	}

	return nil
}

func (u *PaudUseCase) Close() error {
	return u.DB.Close()
}
