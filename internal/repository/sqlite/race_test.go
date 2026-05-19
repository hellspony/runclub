package sqlite_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, clubID, got.ClubID)
	assert.Equal(t, race.Type, got.Type)
	assert.Equal(t, race.Place, got.Place)
	assert.Equal(t, race.Distances, got.Distances)
	assert.Equal(t, race.Name, got.Name)
}

func TestRaceListUpcomingByClub(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewRaceRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "UpClub", -500200)

	from := time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC)

	repo.Create(ctx, &entity.Race{
		ClubID: clubID,
		Date:   time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
		Name:   "October Race",
	})
	repo.Create(ctx, &entity.Race{
		ClubID: clubID,
		Date:   time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
		Name:   "September Race",
	})
	repo.Create(ctx, &entity.Race{
		ClubID: clubID,
		Date:   time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
		Name:   "November Race",
	})

	races, err := repo.ListUpcomingByClub(ctx, clubID, from, to)
	require.NoError(t, err)
	require.Len(t, races, 1)
	assert.Equal(t, "October Race", races[0].Name)
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
	require.NoError(t, err)

	id, err := regRepo.Create(ctx, &entity.RaceRegistration{
		RaceID:   raceID,
		MemberID: memberID,
		Distance: "10k",
	})
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := regRepo.GetByRaceAndMember(ctx, raceID, memberID)
	require.NoError(t, err)
	assert.Equal(t, "10k", got.Distance)
	assert.Equal(t, raceID, got.RaceID)
	assert.Equal(t, memberID, got.MemberID)
}
