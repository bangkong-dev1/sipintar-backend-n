package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/azril773/go-backend/internal/entities"
	"github.com/azril773/go-backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type accessTokenClaims struct {
	entities.AccessTokenClaims
	jwt.RegisteredClaims
}

type refreshTokenClaims struct {
	jwt.RegisteredClaims
}

type AuthUseCase struct {
	UserRepo      *repository.UserRepository
	PaudRepo      *repository.PaudRepository
	AuthRepo      *repository.AuthRepository
	RoleRepo      *repository.RoleRepository
	Logger        *zap.Logger
	AccessSecret  []byte
	RefreshSecret []byte
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

func NewAuthUseCase(userRepo *repository.UserRepository, paudRepo *repository.PaudRepository, authRepo *repository.AuthRepository, roleRepo *repository.RoleRepository, logger *zap.Logger) *AuthUseCase {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if strings.TrimSpace(accessSecret) == "" {
		accessSecret = "local-dev-access-secret-change-me"
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if strings.TrimSpace(refreshSecret) == "" {
		refreshSecret = "local-dev-refresh-secret-change-me"
	}

	accessTTL := parseEnvInt("JWT_ACCESS_TTL_MINUTES", 15)
	refreshTTL := parseEnvInt("JWT_REFRESH_TTL_DAYS", 7)

	return &AuthUseCase{
		UserRepo:      userRepo,
		PaudRepo:      paudRepo,
		AuthRepo:      authRepo,
		RoleRepo:      roleRepo,
		Logger:        logger,
		AccessSecret:  []byte(accessSecret),
		RefreshSecret: []byte(refreshSecret),
		AccessTTL:     time.Duration(accessTTL) * time.Minute,
		RefreshTTL:    time.Duration(refreshTTL) * 24 * time.Hour,
	}
}

func parseEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func (u *AuthUseCase) Login(ctx context.Context, req entities.LoginRequest, requestedSubdomain string) (*entities.AuthResponse, error) {
	user, err := u.UserRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("email atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	claims, err := u.buildAccessClaims(ctx, user)
	if err != nil {
		return nil, err
	}

	requestedSubdomain = strings.TrimSpace(requestedSubdomain)

	if claims.IsCentral {
		if requestedSubdomain != "" {
			return nil, errors.New("akun central tidak dapat login melalui subdomain tenant")
		}
	} else {
		if requestedSubdomain == "" {
			return nil, errors.New("header X-subdomain wajib untuk akun tenant")
		}
		if claims.Subdomain == "" || !strings.EqualFold(claims.Subdomain, requestedSubdomain) {
			return nil, errors.New("akun tidak memiliki akses ke subdomain ini")
		}
	}

	accessToken, err := u.generateAccessToken(user.ID, claims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := u.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &entities.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.AccessTTL.Seconds()),
	}, nil
}

func (u *AuthUseCase) Refresh(ctx context.Context, token string) (*entities.AuthResponse, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("refresh_token wajib diisi")
	}

	claims, err := u.parseRefreshToken(token)
	if err != nil {
		return nil, errors.New("refresh token tidak valid")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, errors.New("refresh token tidak valid")
	}

	hashed := hashToken(token)
	stored, err := u.AuthRepo.GetActiveRefreshToken(ctx, hashed, claims.ID, userID)
	if err != nil {
		return nil, errors.New("refresh token tidak valid atau sudah kadaluarsa")
	}

	if err := u.AuthRepo.RevokeRefreshTokenByID(ctx, stored.ID); err != nil {
		return nil, err
	}

	user, err := u.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	accessClaims, err := u.buildAccessClaims(ctx, user)
	if err != nil {
		return nil, err
	}

	newAccessToken, err := u.generateAccessToken(user.ID, accessClaims)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := u.generateAndStoreRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &entities.AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.AccessTTL.Seconds()),
	}, nil
}

func (u *AuthUseCase) Logout(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("refresh_token wajib diisi")
	}
	return u.AuthRepo.RevokeRefreshTokenByHash(ctx, hashToken(token))
}

func (u *AuthUseCase) ValidateAccessToken(token string) (*entities.AccessTokenClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &accessTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return u.AccessSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(*accessTokenClaims)
	if !ok || !jwtToken.Valid {
		return nil, errors.New("token tidak valid")
	}

	return &claims.AccessTokenClaims, nil
}

func (u *AuthUseCase) buildAccessClaims(ctx context.Context, user *entities.User) (*entities.AccessTokenClaims, error) {
	claims := &entities.AccessTokenClaims{
		UserID:    user.ID,
		PaudID:    user.PaudID,
		RoleID:    user.RoleID,
		IsCentral: user.PaudID == nil,
	}

	if user.PaudID != nil {
		paud, err := u.PaudRepo.GetByID(ctx, *user.PaudID)
		if err != nil {
			return nil, fmt.Errorf("paud user tidak ditemukan")
		}
		claims.Subdomain = paud.Subdomain
		claims.Status = paud.Status
	}

	return claims, nil
}

func (u *AuthUseCase) generateAccessToken(userID uuid.UUID, claims *entities.AccessTokenClaims) (string, error) {
	now := time.Now()
	payload := &accessTokenClaims{
		AccessTokenClaims: *claims,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-backend",
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(u.AccessTTL)),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signed, err := token.SignedString(u.AccessSecret)
	if err != nil {
		return "", fmt.Errorf("gagal generate access token: %w", err)
	}

	return signed, nil
}

func (u *AuthUseCase) generateAndStoreRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	now := time.Now()
	jti := uuid.NewString()
	claims := &refreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-backend",
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(u.RefreshTTL)),
			ID:        jti,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := refreshToken.SignedString(u.RefreshSecret)
	if err != nil {
		return "", fmt.Errorf("gagal generate refresh token: %w", err)
	}

	record := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashToken(signedToken),
		JTI:       jti,
		ExpiresAt: now.Add(u.RefreshTTL),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.AuthRepo.CreateRefreshToken(ctx, record); err != nil {
		return "", err
	}

	return signedToken, nil
}

func (u *AuthUseCase) parseRefreshToken(token string) (*jwt.RegisteredClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &refreshTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return u.RefreshSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(*refreshTokenClaims)
	if !ok || !jwtToken.Valid {
		return nil, errors.New("refresh token invalid")
	}

	return &claims.RegisteredClaims, nil
}

func (u *AuthUseCase) CreateUser(ctx context.Context, email string, password string) (*entities.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal hash password")
	}

	user := &entities.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashed),
	}
	if err := u.UserRepo.Create(ctx, user); err != nil {
		return nil, errors.New("gagal membuat user")
	}
	return user, nil
}

func (u *AuthUseCase) CheckRoleUser(ctx context.Context, claims *entities.AccessTokenClaims, requiredRoles []string) (bool, error) {
	if claims.RoleID == nil {
		return false, errors.New("user tidak memiliki role")
	}
	role, err := u.RoleRepo.GetByID(ctx, *claims.RoleID)
	if err != nil {
		return false, errors.New("role tidak ditemukan")
	}

	if !slices.Contains(requiredRoles, role.Name) {
		return false, nil
	}

	return true, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
