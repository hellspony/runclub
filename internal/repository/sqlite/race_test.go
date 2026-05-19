package sqlite_test

import (
	"testing"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/repository/sqlite"
)

func TestRaceCreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRaceRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "RaceClub", -500100)

	race := &entity.Race{
		ClubID:    clubID,
		Date:      time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
		Type:      "marathon",
		Place:     "City Center",
		Distances: "[\"5k\",\"10k\",\"half\"]",
		Name:      "City Marathon 2025",
	}

	id, err := repo.Create(ctx, race)
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
	if got.Type != race.Type {
		t.Errorf("Type: got %q, want %q", got.Type, race.Type)
	}
	if got.Place != race.Place {
		t.Errorf("Place: got %q, want %q", got.Place, race.Place)
	}
	if got.Distances != race.Distances {
		t.Errorf("Distances: got %q, want %q", got.Distances, race.Distances)
	}
	if got.Name != race.Name {
		t.Errorf("Name: got %q, want %q", got.Name, race.Name)
	}
}

func TestRaceListUpcomingByClub(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRaceRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "UpClub", -500200)

	from := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC)

	// Race within range
	repo.Create(ctx, &entity.Race{
		ClubID: clubID,
		Date:   time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
		Name:   "October Race",
	})
	// Race before range
	repo.Create(ctx, &entity.Race{
		ClubID: clubID,
		Date:   time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
		Name:   "September Race",
	})
	// Race after range
	repo.Create(ctx, &entity.Race{
		ClubID: clubID,
		Date:   time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
		Name:   "November Race",
	})

	races, err := repo.ListUpcomingByClub(ctx, clubID, from, to)
	if err != nil {
		t.Fatalf("ListUpcomingByClub: %v", err)
	}
	if len(races) != 1 {
		t.Fatalf("expected 1 race, got %d", len(races))
	}
	if races[0].Name != "October Race" {
		t.Errorf("Name: got %q, want %q", races[0].Name, "October Race")
	}
}

func TestRaceRegistrationCreate(t *testing.T) {
	db := setupTestDB(t)
	raceRepo := sqlite.NewRaceRepository(db)
	regRepo := sqlite.NewRaceRegistrationRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "RegClub", -500300)
	memberID := mustCreateMember(t, db, "Racer", 8001)

	raceID, err := raceRepo.Create(ctx, &entity.Race{
		ClubID: clubID,
		Date:   time.Date(2025, 11, 20, 0, 0, 0, 0, time.UTC),
		Name:   "Reg Test Race",
	})
	if err != nil {
		t.Fatalf("create race: %v", err)
	}

	id, err := regRepo.Create(ctx, &entity.RaceRegistration{
		RaceID:   raceID,
		MemberID: memberID,
		Distance: "10k",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := regRepo.GetByRaceAndMember(ctx, raceID, memberID)
	if err != nil {
		t.Fatalf("GetByRaceAndMember: %v", err)
	}
	if got.Distance != "10k" {
		t.Errorf("Distance: got %q, want %q", got.Distance, "10k")
	}
	if got.RaceID != raceID {
		t.Errorf("RaceID: got %d, want %d", got.RaceID, raceID)
	}
	if got.MemberID != memberID {
		t.Errorf("MemberID: got %d, want %d", got.MemberID, memberID)
	}
}
