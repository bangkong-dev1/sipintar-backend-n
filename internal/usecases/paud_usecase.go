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
	db     *sqlx.DB
	repo   *repository.PaudRepository
	logger *zap.Logger
}

func NewPaudUseCase(db *sqlx.DB, repo *repository.PaudRepository, logger *zap.Logger) *PaudUseCase {
	return &PaudUseCase{
		db:     db,
		repo:   repo,
		logger: logger.Named("paud"),
	}
}

func (u *PaudUseCase) CreatePaud(ctx context.Context, name, subdomain string) (*entities.Paud, error) {
	if err := u.ValidateSubdomain(subdomain); err != nil {
		return nil, err
	}

	tenant := &entities.Paud{
		ID:        uuid.New(),
		Name:      name,
		Subdomain: subdomain,
		Status:    entities.StatusPending,
	}

	if err := u.repo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	tenant.Status = entities.StatusActive
	if err := u.repo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	u.logger.Info("Tenant created",
		zap.String("tenant_id", tenant.ID.String()),
		zap.String("name", name),
		zap.String("subdomain", subdomain),
	)

	return tenant, nil
}

func (u *PaudUseCase) GetPaud(ctx context.Context, id uuid.UUID) (*entities.Paud, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *PaudUseCase) GetBySubdomain(ctx context.Context, subdomain string) (*entities.Paud, error) {
	return u.repo.GetBySubdomain(ctx, subdomain)
}

func (u *PaudUseCase) UpdatePaud(ctx context.Context, paud *entities.Paud) error {
	if err := u.repo.Update(ctx, paud); err != nil {
		return err
	}
	u.logger.Info("Tenant updated", zap.String("tenant_id", paud.ID.String()))
	return nil
}

func (u *PaudUseCase) DeletePaud(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	u.logger.Info("Tenant deleted", zap.String("tenant_id", id.String()))
	return nil
}

func (u *PaudUseCase) ListPaud(ctx context.Context, page, perPage int) ([]*entities.Paud, int, error) {
	return u.repo.List(ctx, page, perPage)
}

func (u *PaudUseCase) SuspendPaud(ctx context.Context, id uuid.UUID) error {
	tenant, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	tenant.Status = entities.StatusSuspended
	return u.repo.Update(ctx, tenant)
}

func (u *PaudUseCase) ActivatePaud(ctx context.Context, id uuid.UUID) error {
	tenant, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	tenant.Status = entities.StatusActive
	return u.repo.Update(ctx, tenant)
}

func (u *PaudUseCase) GetPaudConn(ctx context.Context, paudID uuid.UUID) (*sql.Conn, error) {
	conn, err := u.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal get connection: %w", err)
	}

	return conn, nil
}

// ResolveFromHost = extract tenant dari host header
func (u *PaudUseCase) ResolveFromHost(ctx context.Context, host string) (*entities.Paud, error) {
	subdomain, err := u.ExtractSubdomain(host)
	if err != nil {
		return nil, err
	}

	tenant, err := u.repo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, fmt.Errorf("tenant tidak ditemukan untuk subdomain: %s", subdomain)
	}

	return tenant, nil
}

// Todo ResolveFromHeader = extract tenant dari header khusus, misal X-Tenant-ID
func (u *PaudUseCase) ExtractFromHeader(c *gin.Context) (string, error) {
	headerName := c.GetHeader("X-Tenant-id")

	subdomain := c.Request.Header.Get(headerName)
	if subdomain == "" {
		return "", fmt.Errorf("no tenant header found: %s", headerName)
	}

	if err := u.ValidateSubdomain(subdomain); err != nil {
		return "", fmt.Errorf("invalid subdomain: %w", err)
	}

	return subdomain, nil
}

// ExtractSubdomain = extract subdomain dari host
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
	return u.db.Close()
}
