package sqlite_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, clubID, got.ClubID)
	assert.Equal(t, location.Name, got.Name)
	assert.Equal(t, location.Address, got.Address)
	assert.Equal(t, location.MapURL, got.MapURL)
	assert.Equal(t, location.Description, got.Description)
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
	require.NoError(t, err)
	require.Len(t, locations, 2)
	assert.Equal(t, "Beach Run", locations[0].Name)
	assert.Equal(t, "Hill Climb", locations[1].Name)
}
