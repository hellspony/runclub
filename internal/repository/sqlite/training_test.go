package sqlite_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"runclub/internal/domain/entity"
	"runclub/internal/repository/sqlite"
)

func TestTrainingCreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewTrainingRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "TrainingClub", -300100)
	locID := mustCreateLocation(t, db, clubID, "Park")

	date := time.Date(2025, 7, 10, 18, 0, 0, 0, time.UTC)
	training := &entity.Training{
		ClubID:      clubID,
		LocationID:  locID,
		Date:        date,
		Duration:    60,
		Status:      entity.TrainingStatusPlanned,
		PhotoFileID: "photo123",
		MessageID:   42,
	}

	id, err := repo.Create(ctx, training)
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	assert.Equal(t, clubID, got.ClubID)
	assert.Equal(t, locID, got.LocationID)
	assert.Equal(t, 60, got.Duration)
	assert.Equal(t, entity.TrainingStatusPlanned, got.Status)
	assert.Equal(t, "photo123", got.PhotoFileID)
	assert.Equal(t, int64(42), got.MessageID)
}

func TestTrainingListByClub(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewTrainingRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "ListClub", -300200)
	locID := mustCreateLocation(t, db, clubID, "Stadium")
	otherClubID := mustCreateClub(t, db, "OtherClub", -300201)

	repo.Create(ctx, &entity.Training{
		ClubID: clubID, LocationID: locID,
		Date: time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC), Status: entity.TrainingStatusPlanned,
	})
	repo.Create(ctx, &entity.Training{
		ClubID: clubID, LocationID: locID,
		Date: time.Date(2025, 8, 2, 0, 0, 0, 0, time.UTC), Status: entity.TrainingStatusPlanned,
	})
	repo.Create(ctx, &entity.Training{
		ClubID: otherClubID, LocationID: locID,
		Date: time.Date(2025, 8, 3, 0, 0, 0, 0, time.UTC), Status: entity.TrainingStatusPlanned,
	})

	trainings, err := repo.ListByClub(ctx, clubID)
	require.NoError(t, err)
	assert.Len(t, trainings, 2)
}

func TestTrainingListByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewTrainingRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "StatusClub", -300300)
	locID := mustCreateLocation(t, db, clubID, "Track")

	repo.Create(ctx, &entity.Training{
		ClubID: clubID, LocationID: locID,
		Date: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC), Status: entity.TrainingStatusPlanned,
	})
	repo.Create(ctx, &entity.Training{
		ClubID: clubID, LocationID: locID,
		Date: time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC), Status: entity.TrainingStatusCompleted,
	})
	repo.Create(ctx, &entity.Training{
		ClubID: clubID, LocationID: locID,
		Date: time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC), Status: entity.TrainingStatusPlanned,
	})

	trainings, err := repo.ListByStatus(ctx, entity.TrainingStatusPlanned)
	require.NoError(t, err)
	assert.Len(t, trainings, 2)
}

func TestTrainingUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := sqlite.NewTrainingRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "UpdClub", -300400)
	locID := mustCreateLocation(t, db, clubID, "Field")

	id, err := repo.Create(ctx, &entity.Training{
		ClubID: clubID, LocationID: locID,
		Date:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
		Status: entity.TrainingStatusPlanned, Duration: 60,
	})
	require.NoError(t, err)

	require.NoError(t, repo.Update(ctx, &entity.Training{
		ID: id, ClubID: clubID, LocationID: locID,
		Date:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
		Status: entity.TrainingStatusCompleted, Duration: 90,
	}))

	got, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, entity.TrainingStatusCompleted, got.Status)
	assert.Equal(t, 90, got.Duration)
}

func TestTrainingTrainerCreate(t *testing.T) {
	db := setupTestDB(t)
	ttRepo := sqlite.NewTrainingTrainerRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "TTClub", -300500)
	locID := mustCreateLocation(t, db, clubID, "Gym")
	memberID := mustCreateMember(t, db, "Trainer", 7001)

	trainingRepo := sqlite.NewTrainingRepository(db)
	trainingID, err := trainingRepo.Create(ctx, &entity.Training{
		ClubID: clubID, LocationID: locID,
		Date:   time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
		Status: entity.TrainingStatusPlanned,
	})
	require.NoError(t, err)

	id, err := ttRepo.Create(ctx, &entity.TrainingTrainer{
		TrainingID: trainingID,
		MemberID:   memberID,
	})
	require.NoError(t, err)
	assert.Positive(t, id)

	trainers, err := ttRepo.ListByTraining(ctx, trainingID)
	require.NoError(t, err)
	require.Len(t, trainers, 1)
	assert.Equal(t, memberID, trainers[0].MemberID)
}

func TestTrainingParticipantCreateAndDelete(t *testing.T) {
	db := setupTestDB(t)
	tpRepo := sqlite.NewTrainingParticipantRepository(db)
	ctx := t.Context()

	clubID := mustCreateClub(t, db, "TPClub", -300600)
	locID := mustCreateLocation(t, db, clubID, "Arena")
	memberID := mustCreateMember(t, db, "Participant", 7002)

	trainingRepo := sqlite.NewTrainingRepository(db)
	trainingID, err := trainingRepo.Create(ctx, &entity.Training{
		ClubID: clubID, LocationID: locID,
		Date:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		Status: entity.TrainingStatusPlanned,
	})
	require.NoError(t, err)

	id, err := tpRepo.Create(ctx, &entity.TrainingParticipant{
		TrainingID: trainingID,
		MemberID:   memberID,
	})
	require.NoError(t, err)
	assert.Positive(t, id)

	got, err := tpRepo.GetByTrainingAndMember(ctx, trainingID, memberID)
	require.NoError(t, err)
	assert.Equal(t, trainingID, got.TrainingID)
	assert.Equal(t, memberID, got.MemberID)

	require.NoError(t, tpRepo.Delete(ctx, id))

	_, err = tpRepo.GetByTrainingAndMember(ctx, trainingID, memberID)
	assert.Error(t, err)
}
