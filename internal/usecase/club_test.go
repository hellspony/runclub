package usecase_test

import (
	"context"
	"testing"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
	"runclub/internal/usecase"
)

// --- mock ---

type mockClubRepo struct {
	clubs  map[int64]*entity.Club
	nextID int64
}

func newMockClubRepo() *mockClubRepo {
	return &mockClubRepo{
		clubs:  make(map[int64]*entity.Club),
		nextID: 1,
	}
}

func (m *mockClubRepo) Create(_ context.Context, club *entity.Club) (int64, error) {
	club.ID = m.nextID
	club.CreatedAt = time.Now()
	club.UpdatedAt = time.Now()
	m.clubs[club.ID] = club
	m.nextID++
	return club.ID, nil
}

func (m *mockClubRepo) GetByID(_ context.Context, id int64) (*entity.Club, error) {
	c, ok := m.clubs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (m *mockClubRepo) GetByTelegramChatID(_ context.Context, chatID int64) (*entity.Club, error) {
	for _, c := range m.clubs {
		if c.TelegramChatID == chatID {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockClubRepo) List(_ context.Context) ([]entity.Club, error) {
	result := make([]entity.Club, 0, len(m.clubs))
	for _, c := range m.clubs {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockClubRepo) Update(_ context.Context, club *entity.Club) error {
	if _, ok := m.clubs[club.ID]; !ok {
		return ErrNotFound
	}
	m.clubs[club.ID] = club
	return nil
}

func (m *mockClubRepo) Delete(_ context.Context, id int64) error {
	delete(m.clubs, id)
	return nil
}

var _ repository.ClubRepository = (*mockClubRepo)(nil)

// --- tests ---

func TestCreateClub(t *testing.T) {
	repo := newMockClubRepo()
	uc := usecase.NewClubUseCase(repo)

	club := &entity.Club{
		Name:           "Runners",
		TelegramChatID: 12345,
	}

	id, err := uc.Create(context.Background(), club)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := uc.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got.Name != "Runners" {
		t.Errorf("expected Name %q, got %q", "Runners", got.Name)
	}
	if got.TelegramChatID != 12345 {
		t.Errorf("expected TelegramChatID 12345, got %d", got.TelegramChatID)
	}
}

func TestGetClubByID(t *testing.T) {
	repo := newMockClubRepo()
	uc := usecase.NewClubUseCase(repo)

	t.Run("existing club", func(t *testing.T) {
		club := &entity.Club{Name: "Alpha", TelegramChatID: 100}
		id, _ := uc.Create(context.Background(), club)

		got, err := uc.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Alpha" {
			t.Errorf("expected Name %q, got %q", "Alpha", got.Name)
		}
	})

	t.Run("non-existent club", func(t *testing.T) {
		_, err := uc.GetByID(context.Background(), 999)
		if err == nil {
			t.Fatal("expected error for non-existent club, got nil")
		}
	})
}

func TestListClubs(t *testing.T) {
	repo := newMockClubRepo()
	uc := usecase.NewClubUseCase(repo)

	_, _ = uc.Create(context.Background(), &entity.Club{Name: "A", TelegramChatID: 1})
	_, _ = uc.Create(context.Background(), &entity.Club{Name: "B", TelegramChatID: 2})

	clubs, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(clubs) != 2 {
		t.Errorf("expected 2 clubs, got %d", len(clubs))
	}

	t.Run("empty list", func(t *testing.T) {
		emptyRepo := newMockClubRepo()
		emptyUC := usecase.NewClubUseCase(emptyRepo)

		emptyClubs, listErr := emptyUC.List(context.Background())
		if listErr != nil {
			t.Fatalf("unexpected error: %v", listErr)
		}
		if len(emptyClubs) != 0 {
			t.Errorf("expected 0 clubs, got %d", len(emptyClubs))
		}
	})
}
