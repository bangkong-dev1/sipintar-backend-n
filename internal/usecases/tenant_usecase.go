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
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type TenantUsecase struct {
	db     *sqlx.DB
	repo   *repository.TenantRepository
	logger *zap.Logger
}

func NewTenantUsecase(db *sqlx.DB, repo *repository.TenantRepository, logger *zap.Logger) *TenantUsecase {
	return &TenantUsecase{
		db:     db,
		repo:   repo,
		logger: logger.Named("tenant"),
	}
}

func (u *TenantUsecase) CreateTenant(ctx context.Context, name, subdomain string) (*entities.Tenant, error) {
	if err := u.ValidateSubdomain(subdomain); err != nil {
		return nil, err
	}

	tenant := &entities.Tenant{
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

func (u *TenantUsecase) GetTenant(ctx context.Context, id uuid.UUID) (*entities.Tenant, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *TenantUsecase) GetBySubdomain(ctx context.Context, subdomain string) (*entities.Tenant, error) {
	return u.repo.GetBySubdomain(ctx, subdomain)
}

func (u *TenantUsecase) UpdateTenant(ctx context.Context, tenant *entities.Tenant) error {
	if err := u.repo.Update(ctx, tenant); err != nil {
		return err
	}
	u.logger.Info("Tenant updated", zap.String("tenant_id", tenant.ID.String()))
	return nil
}

func (u *TenantUsecase) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	u.logger.Info("Tenant deleted", zap.String("tenant_id", id.String()))
	return nil
}

func (u *TenantUsecase) ListTenants(ctx context.Context, page, perPage int) ([]*entities.Tenant, int, error) {
	return u.repo.List(ctx, page, perPage)
}

func (u *TenantUsecase) SuspendTenant(ctx context.Context, id uuid.UUID) error {
	tenant, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	tenant.Status = entities.StatusSuspended
	return u.repo.Update(ctx, tenant)
}

func (u *TenantUsecase) ActivateTenant(ctx context.Context, id uuid.UUID) error {
	tenant, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	tenant.Status = entities.StatusActive
	return u.repo.Update(ctx, tenant)
}

func (u *TenantUsecase) GetTenantConn(ctx context.Context, tenantID uuid.UUID) (*sql.Conn, error) {
	conn, err := u.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal get connection: %w", err)
	}

	return conn, nil
}

// ResolveFromHost = extract tenant dari host header
func (u *TenantUsecase) ResolveFromHost(ctx context.Context, host string) (*entities.Tenant, error) {
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

// ExtractSubdomain = extract subdomain dari host
func (u *TenantUsecase) ExtractSubdomain(host string) (string, error) {
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

func (u *TenantUsecase) ValidateSubdomain(subdomain string) error {
	if len(subdomain) < 3 || len(subdomain) > 50 {
		return errors.New("subdomain harus 3-50 karakter")
	}

	validSubdomain := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !validSubdomain.MatchString(subdomain) {
		return errors.New("subdomain hanya boleh huruf kecil, angka, dan dash (-)")
	}

	return nil
}

func (u *TenantUsecase) Close() error {
	return u.db.Close()
}
