package usecases

import (
	"context"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type RoleUseCase struct {
	DB     *sqlx.DB
	Repo   *repository.RoleRepository
	Logger *zap.Logger
}

func (r *RoleUseCase) ListRoles(ctx context.Context) ([]*entities.Role, error) {
	roles, err := r.Repo.List(ctx)
	if err != nil {
		r.Logger.Error("Failed to list roles", zap.Error(err))
		return nil, err
	}

	return roles, nil
}

func (r *RoleUseCase) GetRoleByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	return r.Repo.GetByID(ctx, id)
}

func (r *RoleUseCase) CreateRole(ctx context.Context, name string) (*entities.Role, error) {
	role := &entities.Role{
		Name: name,
	}
	err := r.Repo.Create(ctx, role)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleUseCase) UpdateRole(ctx context.Context, id uuid.UUID, name string) error {
	role, err := r.Repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	role.Name = name
	return r.Repo.Update(ctx, role)
}

func (r *RoleUseCase) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.Repo.Delete(ctx, id)
}
