package usecases

import (
	"context"
	"fmt"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserUseCase struct {
	DB       *sqlx.DB
	Repo     *repository.UserRepository
	PaudRepo *repository.PaudRepository
	Logger   *zap.Logger
}

func (u *UserUseCase) CreateUser(ctx context.Context, name, email, password string, paudID, roleID *uuid.UUID) (*entities.User, error) {
	existing, _ := u.Repo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, fmt.Errorf("email sudah terdaftar: %s", email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("gagal hash password: %w", err)
	}

	user := &entities.User{
		ID:       uuid.New(),
		PaudID:   paudID,
		RoleID:   roleID,
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := u.Repo.Create(ctx, user); err != nil {
		return nil, err
	}

	u.Logger.Info("user created",
		zap.String("user_id", user.ID.String()),
		zap.String("email", email),
	)

	return user, nil
}

// GetUser = get user by ID
func (u *UserUseCase) GetUser(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	return u.Repo.GetByID(ctx, id)
}

// GetUserByEmail = get user by email
func (u *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return u.Repo.GetByEmail(ctx, email)
}

// GetUsersByPaudSubdomain = get users by paud subdomain (pakai helper paud)
func (u *UserUseCase) GetUsersByPaudSubdomain(ctx context.Context, subdomain string) ([]*entities.User, error) {
	// Get paud by subdomain first
	paud, err := u.PaudRepo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, fmt.Errorf("paud tidak ditemukan: %w", err)
	}

	// Then get users by paud_id
	return u.Repo.GetByPaudID(ctx, paud.ID)
}

// GetUsersByPaudID = get users by paud ID (multi-tenant)
func (u *UserUseCase) GetUsersByPaudID(ctx context.Context, paudID uuid.UUID) ([]*entities.User, error) {
	return u.Repo.GetByPaudID(ctx, paudID)
}

// UpdateUser = update user data
func (u *UserUseCase) UpdateUser(ctx context.Context, user *entities.User) error {
	if err := u.Repo.Update(ctx, user); err != nil {
		return err
	}
	u.Logger.Info("user updated", zap.String("user_id", user.ID.String()))
	return nil
}

// UpdatePassword = update user password
func (u *UserUseCase) UpdatePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	user, err := u.Repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return fmt.Errorf("password lama tidak cocok")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gagal hash password: %w", err)
	}

	if err := u.Repo.UpdatePassword(ctx, id, string(hashedPassword)); err != nil {
		return err
	}

	u.Logger.Info("password updated", zap.String("user_id", id.String()))
	return nil
}

// DeleteUser = delete user
func (u *UserUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := u.Repo.Delete(ctx, id); err != nil {
		return err
	}
	u.Logger.Info("user deleted", zap.String("user_id", id.String()))
	return nil
}

// ListUsers = list all users with pagination
func (u *UserUseCase) ListUsers(ctx context.Context, page, perPage int) ([]*entities.User, int, error) {
	return u.Repo.List(ctx, page, perPage)
}

// ListUsersByPaudID = list users by paud with pagination (multi-tenant)
func (u *UserUseCase) ListUsersByPaudID(ctx context.Context, paudID uuid.UUID, page, perPage int) ([]*entities.User, int, error) {
	return u.Repo.ListByPaudID(ctx, paudID, page, perPage)
}

// ListUsersByPaudSubdomain = list users by paud subdomain with pagination
func (u *UserUseCase) ListUsersByPaudSubdomain(ctx context.Context, subdomain string, page, perPage int) ([]*entities.User, int, error) {
	paud, err := u.PaudRepo.GetBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, 0, fmt.Errorf("paud tidak ditemukan: %w", err)
	}
	return u.Repo.ListByPaudID(ctx, paud.ID, page, perPage)
}

// VerifyPassword = verify user password for login
func (u *UserUseCase) VerifyPassword(ctx context.Context, email, password string) (*entities.User, error) {
	user, err := u.Repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user tidak ditemukan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("password tidak cocok")
	}

	return user, nil
}
