package usecase_test

import (
	"context"
	"testing"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
	"runclub/internal/usecase"

	"golang.org/x/crypto/bcrypt"
)

// --- mocks ---

type mockAdminUserRepo struct {
	users  map[string]*entity.AdminUser
	nextID int64
}

func newMockAdminUserRepo() *mockAdminUserRepo {
	return &mockAdminUserRepo{
		users:  make(map[string]*entity.AdminUser),
		nextID: 1,
	}
}

func (m *mockAdminUserRepo) Create(_ context.Context, user *entity.AdminUser) (int64, error) {
	user.ID = m.nextID
	m.users[user.Username] = user
	m.nextID++
	return user.ID, nil
}

func (m *mockAdminUserRepo) GetByID(_ context.Context, id int64) (*entity.AdminUser, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockAdminUserRepo) GetByUsername(_ context.Context, username string) (*entity.AdminUser, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (m *mockAdminUserRepo) List(_ context.Context) ([]entity.AdminUser, error) {
	var result []entity.AdminUser
	for _, u := range m.users {
		result = append(result, *u)
	}
	return result, nil
}

func (m *mockAdminUserRepo) Update(_ context.Context, user *entity.AdminUser) error {
	m.users[user.Username] = user
	return nil
}

func (m *mockAdminUserRepo) Delete(_ context.Context, id int64) error {
	for username, u := range m.users {
		if u.ID == id {
			delete(m.users, username)
			return nil
		}
	}
	return nil
}

var _ repository.AdminUserRepository = (*mockAdminUserRepo)(nil)

type mockAdminUserClubRepo struct {
	assignments map[int64]map[int64]bool // adminUserID -> clubID -> true
}

func newMockAdminUserClubRepo() *mockAdminUserClubRepo {
	return &mockAdminUserClubRepo{
		assignments: make(map[int64]map[int64]bool),
	}
}

func (m *mockAdminUserClubRepo) Add(_ context.Context, adminUserID, clubID int64) error {
	if m.assignments[adminUserID] == nil {
		m.assignments[adminUserID] = make(map[int64]bool)
	}
	m.assignments[adminUserID][clubID] = true
	return nil
}

func (m *mockAdminUserClubRepo) Remove(_ context.Context, adminUserID, clubID int64) error {
	delete(m.assignments[adminUserID], clubID)
	return nil
}

func (m *mockAdminUserClubRepo) ListByAdminUser(_ context.Context, adminUserID int64) ([]int64, error) {
	var ids []int64
	for id := range m.assignments[adminUserID] {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *mockAdminUserClubRepo) ListByClub(_ context.Context, clubID int64) ([]int64, error) {
	var ids []int64
	for uid, clubs := range m.assignments {
		if clubs[clubID] {
			ids = append(ids, uid)
		}
	}
	return ids, nil
}

var _ repository.AdminUserClubRepository = (*mockAdminUserClubRepo)(nil)

const testJWTSecret = "test-secret"

// --- tests ---

func TestLogin_Success(t *testing.T) {
	repo := newMockAdminUserRepo()
	uc := usecase.NewAuthUseCase(repo, newMockAdminUserClubRepo(), testJWTSecret)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.Create(context.Background(), &entity.AdminUser{
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         entity.AdminRoleSuperAdmin,
	})

	token, err := uc.Login(context.Background(), "admin", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := uc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if claims.Username != "admin" {
		t.Errorf("expected username %q, got %q", "admin", claims.Username)
	}
	if claims.Role != entity.AdminRoleSuperAdmin {
		t.Errorf("expected role %q, got %q", entity.AdminRoleSuperAdmin, claims.Role)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := newMockAdminUserRepo()
	uc := usecase.NewAuthUseCase(repo, newMockAdminUserClubRepo(), testJWTSecret)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo.Create(context.Background(), &entity.AdminUser{
		Username:     "admin",
		PasswordHash: string(hash),
	})

	_, err := uc.Login(context.Background(), "admin", "wrongpassword")
	if err == nil {
		t.Fatal("expected error for invalid password, got nil")
	}

	t.Run("non-existent user", func(t *testing.T) {
		_, loginErr := uc.Login(context.Background(), "nobody", "whatever")
		if loginErr == nil {
			t.Fatal("expected error for non-existent user")
		}
	})
}

func TestValidateToken_Valid(t *testing.T) {
	repo := newMockAdminUserRepo()
	uc := usecase.NewAuthUseCase(repo, newMockAdminUserClubRepo(), testJWTSecret)

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	repo.Create(context.Background(), &entity.AdminUser{
		Username:     "testuser",
		PasswordHash: string(hash),
		Role:         entity.AdminRoleAdmin,
	})

	token, _ := uc.Login(context.Background(), "testuser", "pass")
	claims, err := uc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Username != "testuser" {
		t.Errorf("expected username %q, got %q", "testuser", claims.Username)
	}
	if claims.Role != entity.AdminRoleAdmin {
		t.Errorf("expected role %q, got %q", entity.AdminRoleAdmin, claims.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	repo := newMockAdminUserRepo()
	uc := usecase.NewAuthUseCase(repo, newMockAdminUserClubRepo(), testJWTSecret)

	t.Run("garbage token", func(t *testing.T) {
		_, err := uc.ValidateToken(context.Background(), "not-a-real-token")
		if err == nil {
			t.Fatal("expected error for invalid token")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		otherUC := usecase.NewAuthUseCase(repo, newMockAdminUserClubRepo(), "other-secret")
		hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
		repo.Create(context.Background(), &entity.AdminUser{
			Username:     "other",
			PasswordHash: string(hash),
		})
		token, _ := otherUC.Login(context.Background(), "other", "pass")

		_, err := uc.ValidateToken(context.Background(), token)
		if err == nil {
			t.Fatal("expected error for token signed with wrong secret")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := uc.ValidateToken(context.Background(), "")
		if err == nil {
			t.Fatal("expected error for empty token")
		}
	})
}

func TestSeedAdmin(t *testing.T) {
	repo := newMockAdminUserRepo()
	uc := usecase.NewAuthUseCase(repo, newMockAdminUserClubRepo(), testJWTSecret)

	t.Run("creates new admin with role", func(t *testing.T) {
		err := uc.SeedAdmin(context.Background(), "superadmin", "securepass", "superadmin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		user, err := repo.GetByUsername(context.Background(), "superadmin")
		if err != nil {
			t.Fatalf("admin not found after seed: %v", err)
		}
		if user.Username != "superadmin" {
			t.Errorf("expected username %q, got %q", "superadmin", user.Username)
		}
		if user.Role != entity.AdminRoleSuperAdmin {
			t.Errorf("expected role %q, got %q", entity.AdminRoleSuperAdmin, user.Role)
		}
		if cmpErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("securepass")); cmpErr != nil {
			t.Error("seeded password does not match")
		}
	})

	t.Run("idempotent - does not overwrite existing admin", func(t *testing.T) {
		err := uc.SeedAdmin(context.Background(), "superadmin", "newpass", "admin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		user, _ := repo.GetByUsername(context.Background(), "superadmin")
		if cmpErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("securepass")); cmpErr != nil {
			t.Error("original password was overwritten by second seed")
		}
	})
}
