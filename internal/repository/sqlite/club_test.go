package sqlite_test

import (
	"testing"

	"runclub/internal/domain/entity"
	"runclub/internal/repository/sqlite"
)

func TestClubCreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubRepository(db)
	ctx := t.Context()

	club := &entity.Club{
		Name:              "Run Club",
		TelegramChatID:    -1001234567890,
		WelcomeEnabled:    true,
		BirthdayEnabled:   true,
		RaceNotifyEnabled: false,
	}

	id, err := repo.Create(ctx, club)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Name != club.Name {
		t.Errorf("Name: got %q, want %q", got.Name, club.Name)
	}
	if got.TelegramChatID != club.TelegramChatID {
		t.Errorf("TelegramChatID: got %d, want %d", got.TelegramChatID, club.TelegramChatID)
	}
	if got.WelcomeEnabled != club.WelcomeEnabled {
		t.Errorf("WelcomeEnabled: got %v, want %v", got.WelcomeEnabled, club.WelcomeEnabled)
	}
	if got.BirthdayEnabled != club.BirthdayEnabled {
		t.Errorf("BirthdayEnabled: got %v, want %v", got.BirthdayEnabled, club.BirthdayEnabled)
	}
	if got.RaceNotifyEnabled != club.RaceNotifyEnabled {
		t.Errorf("RaceNotifyEnabled: got %v, want %v", got.RaceNotifyEnabled, club.RaceNotifyEnabled)
	}
	if got.ID != id {
		t.Errorf("ID: got %d, want %d", got.ID, id)
	}
}

func TestClubGetByTelegramChatID(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubRepository(db)
	ctx := t.Context()

	chatID := int64(-100999888777)
	id, err := repo.Create(ctx, &entity.Club{
		Name:           "FindMe",
		TelegramChatID: chatID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByTelegramChatID(ctx, chatID)
	if err != nil {
		t.Fatalf("GetByTelegramChatID: %v", err)
	}
	if got.ID != id {
		t.Errorf("ID: got %d, want %d", got.ID, id)
	}
	if got.Name != "FindMe" {
		t.Errorf("Name: got %q, want %q", got.Name, "FindMe")
	}
}

func TestClubList(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubRepository(db)
	ctx := t.Context()

	for i, name := range []string{"Alpha", "Beta", "Gamma"} {
		_, err := repo.Create(ctx, &entity.Club{
			Name:           name,
			TelegramChatID: int64(-100 + i),
		})
		if err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}

	clubs, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clubs) != 3 {
		t.Fatalf("expected 3 clubs, got %d", len(clubs))
	}
	// List returns ORDER BY id, so order should match insertion order
	if clubs[0].Name != "Alpha" || clubs[1].Name != "Beta" || clubs[2].Name != "Gamma" {
		t.Errorf("order: got %q, %q, %q; want Alpha, Beta, Gamma",
			clubs[0].Name, clubs[1].Name, clubs[2].Name)
	}
}

func TestClubUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubRepository(db)
	ctx := t.Context()

	id, err := repo.Create(ctx, &entity.Club{
		Name:              "Original",
		TelegramChatID:    -100111,
		WelcomeEnabled:    true,
		BirthdayEnabled:   true,
		RaceNotifyEnabled: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated := &entity.Club{
		ID:                id,
		Name:              "Updated",
		TelegramChatID:    -100222,
		WelcomeEnabled:    false,
		BirthdayEnabled:   false,
		RaceNotifyEnabled: false,
	}
	if err = repo.Update(ctx, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "Updated" {
		t.Errorf("Name: got %q, want %q", got.Name, "Updated")
	}
	if got.WelcomeEnabled != false {
		t.Errorf("WelcomeEnabled: got %v, want false", got.WelcomeEnabled)
	}
	if got.BirthdayEnabled != false {
		t.Errorf("BirthdayEnabled: got %v, want false", got.BirthdayEnabled)
	}
	if got.RaceNotifyEnabled != false {
		t.Errorf("RaceNotifyEnabled: got %v, want false", got.RaceNotifyEnabled)
	}
}

func TestClubDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubRepository(db)
	ctx := t.Context()

	id, err := repo.Create(ctx, &entity.Club{
		Name:           "ToDelete",
		TelegramChatID: -100333,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err = repo.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(ctx, id)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
