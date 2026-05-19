package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	go_mock "go.uber.org/mock/gomock"

	"runclub/internal/domain/entity"
	"runclub/internal/mocks"
	"runclub/internal/usecase"

	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "test-secret"

func TestLogin_Success(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAdminUserRepository(ctrl)
	clubRepo := mocks.NewMockAdminUserClubRepository(ctrl)
	uc := usecase.NewAuthUseCase(repo, clubRepo, testJWTSecret)

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repo.EXPECT().
		GetByUsername(go_mock.Any(), "admin").
		Return(&entity.AdminUser{
			ID:           1,
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         entity.AdminRoleSuperAdmin,
		}, nil)

	token, err := uc.Login(context.Background(), "admin", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := uc.ValidateToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, "admin", claims.Username)
	assert.Equal(t, entity.AdminRoleSuperAdmin, claims.Role)
}

func TestLogin_InvalidPassword(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAdminUserRepository(ctrl)
	clubRepo := mocks.NewMockAdminUserClubRepository(ctrl)
	uc := usecase.NewAuthUseCase(repo, clubRepo, testJWTSecret)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo.EXPECT().
		GetByUsername(go_mock.Any(), "admin").
		Return(&entity.AdminUser{
			ID:           1,
			Username:     "admin",
			PasswordHash: string(hash),
		}, nil)

	_, err := uc.Login(context.Background(), "admin", "wrongpassword")
	require.Error(t, err)

	t.Run("non-existent user", func(t *testing.T) {
		repo.EXPECT().
			GetByUsername(go_mock.Any(), "nobody").
			Return(nil, assert.AnError)

		_, loginErr := uc.Login(context.Background(), "nobody", "whatever")
		assert.Error(t, loginErr)
	})
}

func TestValidateToken_Valid(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAdminUserRepository(ctrl)
	clubRepo := mocks.NewMockAdminUserClubRepository(ctrl)
	uc := usecase.NewAuthUseCase(repo, clubRepo, testJWTSecret)

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	repo.EXPECT().
		GetByUsername(go_mock.Any(), "testuser").
		Return(&entity.AdminUser{
			ID:           1,
			Username:     "testuser",
			PasswordHash: string(hash),
			Role:         entity.AdminRoleAdmin,
		}, nil)

	token, err := uc.Login(context.Background(), "testuser", "pass")
	require.NoError(t, err)

	claims, err := uc.ValidateToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, entity.AdminRoleAdmin, claims.Role)
}

func TestValidateToken_Invalid(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAdminUserRepository(ctrl)
	clubRepo := mocks.NewMockAdminUserClubRepository(ctrl)
	uc := usecase.NewAuthUseCase(repo, clubRepo, testJWTSecret)

	t.Run("garbage token", func(t *testing.T) {
		_, err := uc.ValidateToken(context.Background(), "not-a-real-token")
		assert.Error(t, err)
	})

	t.Run("wrong secret", func(t *testing.T) {
		otherRepo := mocks.NewMockAdminUserRepository(ctrl)
		otherClubRepo := mocks.NewMockAdminUserClubRepository(ctrl)
		otherUC := usecase.NewAuthUseCase(otherRepo, otherClubRepo, "other-secret")

		hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
		otherRepo.EXPECT().
			GetByUsername(go_mock.Any(), "other").
			Return(&entity.AdminUser{
				ID:           2,
				Username:     "other",
				PasswordHash: string(hash),
			}, nil)

		token, _ := otherUC.Login(context.Background(), "other", "pass")
		_, err := uc.ValidateToken(context.Background(), token)
		assert.Error(t, err)
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := uc.ValidateToken(context.Background(), "")
		assert.Error(t, err)
	})
}

func TestSeedAdmin(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockAdminUserRepository(ctrl)
	clubRepo := mocks.NewMockAdminUserClubRepository(ctrl)
	uc := usecase.NewAuthUseCase(repo, clubRepo, testJWTSecret)

	t.Run("creates new admin with role", func(t *testing.T) {
		repo.EXPECT().
			GetByUsername(go_mock.Any(), "superadmin").
			Return(nil, assert.AnError)
		repo.EXPECT().
			Create(go_mock.Any(), go_mock.Any()).
			Return(int64(1), nil)

		err := uc.SeedAdmin(context.Background(), "superadmin", "securepass", "superadmin")
		require.NoError(t, err)
	})

	t.Run("idempotent - does not overwrite when role matches", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("securepass"), bcrypt.DefaultCost)
		repo.EXPECT().
			GetByUsername(go_mock.Any(), "superadmin").
			Return(&entity.AdminUser{
				ID:           1,
				Username:     "superadmin",
				PasswordHash: string(hash),
				Role:         entity.AdminRoleSuperAdmin,
			}, nil)

		err := uc.SeedAdmin(context.Background(), "superadmin", "newpass", "superadmin")
		require.NoError(t, err)
	})

	t.Run("updates role when it differs", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("securepass"), bcrypt.DefaultCost)
		repo.EXPECT().
			GetByUsername(go_mock.Any(), "superadmin").
			Return(&entity.AdminUser{
				ID:           1,
				Username:     "superadmin",
				PasswordHash: string(hash),
				Role:         entity.AdminRoleSuperAdmin,
			}, nil)
		repo.EXPECT().
			Update(go_mock.Any(), go_mock.Any()).
			Return(nil)

		err := uc.SeedAdmin(context.Background(), "superadmin", "newpass", "admin")
		require.NoError(t, err)
	})
}
