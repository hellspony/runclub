package usecase

import (
	"context"
	"errors"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type TokenClaims struct {
	UserID   int64
	Username string
	Role     entity.AdminRole
}

type AuthUseCase interface {
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	SeedAdmin(ctx context.Context, username, password, role string) error
	CreateAdminUser(ctx context.Context, username, password, role string) (int64, error)
	ListAdminUsers(ctx context.Context) ([]entity.AdminUser, error)
	DeleteAdminUser(ctx context.Context, id int64) error
	AssignClub(ctx context.Context, adminUserID, clubID int64) error
	UnassignClub(ctx context.Context, adminUserID, clubID int64) error
	GetAdminUserClubs(ctx context.Context, adminUserID int64) ([]int64, error)
}

type authUseCase struct {
	adminRepo     repository.AdminUserRepository
	adminClubRepo repository.AdminUserClubRepository
	jwtSecret     string
}

func NewAuthUseCase(
	adminRepo repository.AdminUserRepository,
	adminClubRepo repository.AdminUserClubRepository,
	jwtSecret string,
) AuthUseCase {
	return &authUseCase{
		adminRepo:     adminRepo,
		adminClubRepo: adminClubRepo,
		jwtSecret:     jwtSecret,
	}
}

const tokenExpiryHours = 24

type authClaims struct {
	jwt.RegisteredClaims

	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func (uc *authUseCase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := uc.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	now := time.Now()
	claims := authClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpiryHours * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (uc *authUseCase) ValidateToken(_ context.Context, tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &authClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(uc.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*authClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &TokenClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     entity.AdminRole(claims.Role),
	}, nil
}

func (uc *authUseCase) SeedAdmin(ctx context.Context, username, password, role string) error {
	existing, err := uc.adminRepo.GetByUsername(ctx, username)
	if err == nil && existing != nil {
		// Update role if it changed (e.g., from admin to superadmin via env)
		if existing.Role != entity.AdminRole(role) {
			existing.Role = entity.AdminRole(role)
			return uc.adminRepo.Update(ctx, existing)
		}
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &entity.AdminUser{
		Username:     username,
		PasswordHash: string(hash),
		Role:         entity.AdminRole(role),
	}

	_, err = uc.adminRepo.Create(ctx, admin)
	return err
}

func (uc *authUseCase) CreateAdminUser(ctx context.Context, username, password, role string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	user := &entity.AdminUser{
		Username:     username,
		PasswordHash: string(hash),
		Role:         entity.AdminRole(role),
	}

	return uc.adminRepo.Create(ctx, user)
}

func (uc *authUseCase) ListAdminUsers(ctx context.Context) ([]entity.AdminUser, error) {
	return uc.adminRepo.List(ctx)
}

func (uc *authUseCase) DeleteAdminUser(ctx context.Context, id int64) error {
	return uc.adminRepo.Delete(ctx, id)
}

func (uc *authUseCase) AssignClub(ctx context.Context, adminUserID, clubID int64) error {
	return uc.adminClubRepo.Add(ctx, adminUserID, clubID)
}

func (uc *authUseCase) UnassignClub(ctx context.Context, adminUserID, clubID int64) error {
	return uc.adminClubRepo.Remove(ctx, adminUserID, clubID)
}

func (uc *authUseCase) GetAdminUserClubs(ctx context.Context, adminUserID int64) ([]int64, error) {
	return uc.adminClubRepo.ListByAdminUser(ctx, adminUserID)
}
