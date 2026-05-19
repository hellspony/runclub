package sqlite_test

import (
	"testing"

	"runclub/internal/domain/entity"
	"runclub/internal/repository/sqlite"
)

func TestLocationCreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLocationRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "LocClub", -400100)

	location := &entity.Location{
		ClubID:      clubID,
		Name:        "Central Park",
		Address:     "123 Main St",
		MapURL:      "https://maps.example.com/1",
		Description: "Nice running route",
	}

	id, err := repo.Create(ctx, location)
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

	if got.ClubID != clubID {
		t.Errorf("ClubID: got %d, want %d", got.ClubID, clubID)
	}
	if got.Name != location.Name {
		t.Errorf("Name: got %q, want %q", got.Name, location.Name)
	}
	if got.Address != location.Address {
		t.Errorf("Address: got %q, want %q", got.Address, location.Address)
	}
	if got.MapURL != location.MapURL {
		t.Errorf("MapURL: got %q, want %q", got.MapURL, location.MapURL)
	}
	if got.Description != location.Description {
		t.Errorf("Description: got %q, want %q", got.Description, location.Description)
	}
}

func TestLocationListByClub(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewLocationRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "LocListClub", -400200)
	otherClubID := mustCreateClub(t, db, "LocOtherClub", -400201)

	repo.Create(ctx, &entity.Location{ClubID: clubID, Name: "Beach Run"})
	repo.Create(ctx, &entity.Location{ClubID: clubID, Name: "Hill Climb"})
	repo.Create(ctx, &entity.Location{ClubID: otherClubID, Name: "Downtown"})

	locations, err := repo.ListByClub(ctx, clubID)
	if err != nil {
		t.Fatalf("ListByClub: %v", err)
	}
	if len(locations) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(locations))
	}
	// Ordered by name
	if locations[0].Name != "Beach Run" {
		t.Errorf("first location: got %q, want %q", locations[0].Name, "Beach Run")
	}
	if locations[1].Name != "Hill Climb" {
		t.Errorf("second location: got %q, want %q", locations[1].Name, "Hill Climb")
	}
}
