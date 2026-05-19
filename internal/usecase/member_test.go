package usecase_test

import (
	"context"
	"errors"
	"testing"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
	"runclub/internal/usecase"
)

// --- mocks ---

type mockMemberRepo struct {
	members map[int64]*entity.Member
	nextID  int64
}

func newMockMemberRepo() *mockMemberRepo {
	return &mockMemberRepo{
		members: make(map[int64]*entity.Member),
		nextID:  1,
	}
}

func (m *mockMemberRepo) Create(_ context.Context, member *entity.Member) (int64, error) {
	member.ID = m.nextID
	m.members[member.ID] = member
	m.nextID++
	return member.ID, nil
}

func (m *mockMemberRepo) GetByID(_ context.Context, id int64) (*entity.Member, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, ErrNotFound
	}
	return member, nil
}

func (m *mockMemberRepo) GetByTelegramID(_ context.Context, tid int64) (*entity.Member, error) {
	for _, m := range m.members {
		if m.TelegramID == tid {
			return m, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockMemberRepo) ListByClub(_ context.Context, _ int64) ([]entity.Member, error) {
	return nil, nil
}

func (m *mockMemberRepo) ListTrainersByClub(_ context.Context, _ int64) ([]entity.Member, error) {
	return nil, nil
}

func (m *mockMemberRepo) ListBirthdayOn(_ context.Context, _, _ int) ([]entity.Member, error) {
	return nil, nil
}

func (m *mockMemberRepo) ListOrphansOlderThan(_ context.Context, _ int) ([]entity.Member, error) {
	return nil, nil
}

func (m *mockMemberRepo) Update(_ context.Context, member *entity.Member) error {
	if _, ok := m.members[member.ID]; !ok {
		return ErrNotFound
	}
	m.members[member.ID] = member
	return nil
}

func (m *mockMemberRepo) Delete(_ context.Context, id int64) error {
	delete(m.members, id)
	return nil
}

var _ repository.MemberRepository = (*mockMemberRepo)(nil)

type mockClubMemberRepo struct {
	records map[clubMemberKey]*entity.ClubMember
	nextID  int64
}

type clubMemberKey struct {
	ClubID, MemberID int64
}

func newMockClubMemberRepo() *mockClubMemberRepo {
	return &mockClubMemberRepo{
		records: make(map[clubMemberKey]*entity.ClubMember),
		nextID:  1,
	}
}

func (m *mockClubMemberRepo) Create(_ context.Context, cm *entity.ClubMember) (int64, error) {
	key := clubMemberKey{cm.ClubID, cm.MemberID}
	cm.ID = m.nextID
	m.records[key] = cm
	m.nextID++
	return cm.ID, nil
}

func (m *mockClubMemberRepo) GetByClubAndMember(_ context.Context, clubID, memberID int64) (*entity.ClubMember, error) {
	cm, ok := m.records[clubMemberKey{clubID, memberID}]
	if !ok {
		return nil, ErrNotFound
	}
	return cm, nil
}

func (m *mockClubMemberRepo) ListClubsByMember(_ context.Context, _ int64) ([]entity.ClubMember, error) {
	return nil, nil
}

func (m *mockClubMemberRepo) ListTrainerClubs(_ context.Context, _ int64) ([]entity.ClubMember, error) {
	return nil, nil
}

func (m *mockClubMemberRepo) UpdateRole(_ context.Context, clubID, memberID int64, role entity.MemberRole) error {
	key := clubMemberKey{clubID, memberID}
	cm, ok := m.records[key]
	if !ok {
		return ErrNotFound
	}
	cm.Role = role
	return nil
}

func (m *mockClubMemberRepo) Delete(_ context.Context, clubID, memberID int64) error {
	delete(m.records, clubMemberKey{clubID, memberID})
	return nil
}

var _ repository.ClubMemberRepository = (*mockClubMemberRepo)(nil)

// --- tests ---

func TestRegisterOrGet_NewMember(t *testing.T) {
	memberRepo := newMockMemberRepo()
	clubMemberRepo := newMockClubMemberRepo()
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	member, err := uc.RegisterOrGet(context.Background(), 111, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if member.ID == 0 {
		t.Fatal("expected non-zero ID for new member")
	}
	if member.TelegramID != 111 {
		t.Errorf("expected TelegramID 111, got %d", member.TelegramID)
	}
	if member.FIO != "alice" {
		t.Errorf("expected FIO %q, got %q", "alice", member.FIO)
	}
	if member.TelegramUsername != "alice" {
		t.Errorf("expected TelegramUsername %q, got %q", "alice", member.TelegramUsername)
	}
}

func TestRegisterOrGet_ExistingMember(t *testing.T) {
	memberRepo := newMockMemberRepo()
	clubMemberRepo := newMockClubMemberRepo()
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	// Register first time
	first, _ := uc.RegisterOrGet(context.Background(), 222, "bob")

	// Register again with same TelegramID
	second, err := uc.RegisterOrGet(context.Background(), 222, "bob_updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if second.ID != first.ID {
		t.Errorf("expected same ID %d, got %d", first.ID, second.ID)
	}
	if second.TelegramID != 222 {
		t.Errorf("expected TelegramID 222, got %d", second.TelegramID)
	}
}

func TestAddToClub(t *testing.T) {
	memberRepo := newMockMemberRepo()
	clubMemberRepo := newMockClubMemberRepo()
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	member, _ := uc.RegisterOrGet(context.Background(), 333, "charlie")

	err := uc.AddToClub(context.Background(), 10, member.ID, entity.RoleMember)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cm, err := uc.GetClubMember(context.Background(), 10, member.ID)
	if err != nil {
		t.Fatalf("GetClubMember returned error: %v", err)
	}
	if cm.Role != entity.RoleMember {
		t.Errorf("expected role %q, got %q", entity.RoleMember, cm.Role)
	}
	if cm.ClubID != 10 {
		t.Errorf("expected ClubID 10, got %d", cm.ClubID)
	}
}

func TestUpdateRole(t *testing.T) {
	memberRepo := newMockMemberRepo()
	clubMemberRepo := newMockClubMemberRepo()
	uc := usecase.NewMemberUseCase(memberRepo, clubMemberRepo)

	member, _ := uc.RegisterOrGet(context.Background(), 444, "dave")
	_ = uc.AddToClub(context.Background(), 20, member.ID, entity.RoleMember)

	err := uc.UpdateRole(context.Background(), 20, member.ID, entity.RoleTrainer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cm, _ := uc.GetClubMember(context.Background(), 20, member.ID)
	if cm.Role != entity.RoleTrainer {
		t.Errorf("expected role %q, got %q", entity.RoleTrainer, cm.Role)
	}

	t.Run("non-existent membership", func(t *testing.T) {
		updateErr := uc.UpdateRole(context.Background(), 999, 999, entity.RoleAdmin)
		if updateErr == nil {
			t.Fatal("expected error for non-existent membership")
		}
	})
}

// ErrNotFound is a shared sentinel used by all mock repos in this package.
var ErrNotFound = errors.New("not found")
