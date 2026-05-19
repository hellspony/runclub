package sqlite_test

import (
	"testing"
	"time"

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
	if got.LocationID != locID {
		t.Errorf("LocationID: got %d, want %d", got.LocationID, locID)
	}
	if got.Duration != 60 {
		t.Errorf("Duration: got %d, want 60", got.Duration)
	}
	if got.Status != entity.TrainingStatusPlanned {
		t.Errorf("Status: got %q, want %q", got.Status, entity.TrainingStatusPlanned)
	}
	if got.PhotoFileID != "photo123" {
		t.Errorf("PhotoFileID: got %q, want %q", got.PhotoFileID, "photo123")
	}
	if got.MessageID != 42 {
		t.Errorf("MessageID: got %d, want 42", got.MessageID)
	}
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
	if err != nil {
		t.Fatalf("ListByClub: %v", err)
	}
	if len(trainings) != 2 {
		t.Fatalf("expected 2 trainings, got %d", len(trainings))
	}
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
	if err != nil {
		t.Fatalf("ListByStatus: %v", err)
	}
	if len(trainings) != 2 {
		t.Fatalf("expected 2 planned trainings, got %d", len(trainings))
	}
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
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err = repo.Update(ctx, &entity.Training{
		ID: id, ClubID: clubID, LocationID: locID,
		Date:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
		Status: entity.TrainingStatusCompleted, Duration: 90,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Status != entity.TrainingStatusCompleted {
		t.Errorf("Status: got %q, want %q", got.Status, entity.TrainingStatusCompleted)
	}
	if got.Duration != 90 {
		t.Errorf("Duration: got %d, want 90", got.Duration)
	}
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
	if err != nil {
		t.Fatalf("create training: %v", err)
	}

	id, err := ttRepo.Create(ctx, &entity.TrainingTrainer{
		TrainingID: trainingID,
		MemberID:   memberID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	trainers, err := ttRepo.ListByTraining(ctx, trainingID)
	if err != nil {
		t.Fatalf("ListByTraining: %v", err)
	}
	if len(trainers) != 1 {
		t.Fatalf("expected 1 trainer, got %d", len(trainers))
	}
	if trainers[0].MemberID != memberID {
		t.Errorf("MemberID: got %d, want %d", trainers[0].MemberID, memberID)
	}
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
	if err != nil {
		t.Fatalf("create training: %v", err)
	}

	id, err := tpRepo.Create(ctx, &entity.TrainingParticipant{
		TrainingID: trainingID,
		MemberID:   memberID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := tpRepo.GetByTrainingAndMember(ctx, trainingID, memberID)
	if err != nil {
		t.Fatalf("GetByTrainingAndMember: %v", err)
	}
	if got.TrainingID != trainingID {
		t.Errorf("TrainingID: got %d, want %d", got.TrainingID, trainingID)
	}
	if got.MemberID != memberID {
		t.Errorf("MemberID: got %d, want %d", got.MemberID, memberID)
	}

	if err = tpRepo.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = tpRepo.GetByTrainingAndMember(ctx, trainingID, memberID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}
