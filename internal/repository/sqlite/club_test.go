package sqlite_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, club.Name, got.Name)
	assert.Equal(t, club.TelegramChatID, got.TelegramChatID)
	assert.Equal(t, club.WelcomeEnabled, got.WelcomeEnabled)
	assert.Equal(t, club.BirthdayEnabled, got.BirthdayEnabled)
	assert.Equal(t, club.RaceNotifyEnabled, got.RaceNotifyEnabled)
	assert.Equal(t, id, got.ID)
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
	require.NoError(t, err)

	got, err := repo.GetByTelegramChatID(ctx, chatID)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "FindMe", got.Name)
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
		require.NoError(t, err, "Create %s", name)
	}

	clubs, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, clubs, 3)
	assert.Equal(t, "Alpha", clubs[0].Name)
	assert.Equal(t, "Beta", clubs[1].Name)
	assert.Equal(t, "Gamma", clubs[2].Name)
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
	require.NoError(t, err)

	updated := &entity.Club{
		ID:                id,
		Name:              "Updated",
		TelegramChatID:    -100222,
		WelcomeEnabled:    false,
		BirthdayEnabled:   false,
		RaceNotifyEnabled: false,
	}
	require.NoError(t, repo.Update(ctx, updated))

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "Updated", got.Name)
	assert.False(t, got.WelcomeEnabled)
	assert.False(t, got.BirthdayEnabled)
	assert.False(t, got.RaceNotifyEnabled)
}

func TestClubDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewClubRepository(db)
	ctx := t.Context()

	id, err := repo.Create(ctx, &entity.Club{
		Name:           "ToDelete",
		TelegramChatID: -100333,
	})
	require.NoError(t, err)

	require.NoError(t, repo.Delete(ctx, id))

	_, err = repo.GetByID(ctx, id)
	assert.Error(t, err)
}
