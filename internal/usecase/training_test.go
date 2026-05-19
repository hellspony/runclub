package usecase_test

import (
	"context"
	"testing"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
	"runclub/internal/usecase"
)

// --- mocks ---

type mockTrainingRepo struct {
	trainings map[int64]*entity.Training
	nextID    int64
}

func newMockTrainingRepo() *mockTrainingRepo {
	return &mockTrainingRepo{
		trainings: make(map[int64]*entity.Training),
		nextID:    1,
	}
}

func (m *mockTrainingRepo) Create(_ context.Context, t *entity.Training) (int64, error) {
	t.ID = m.nextID
	m.trainings[t.ID] = t
	m.nextID++
	return t.ID, nil
}

func (m *mockTrainingRepo) GetByID(_ context.Context, id int64) (*entity.Training, error) {
	t, ok := m.trainings[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (m *mockTrainingRepo) ListByClub(_ context.Context, clubID int64) ([]entity.Training, error) {
	var result []entity.Training
	for _, t := range m.trainings {
		if t.ClubID == clubID {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (m *mockTrainingRepo) ListByStatus(_ context.Context, status entity.TrainingStatus) ([]entity.Training, error) {
	var result []entity.Training
	for _, t := range m.trainings {
		if t.Status == status {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (m *mockTrainingRepo) Update(_ context.Context, t *entity.Training) error {
	if _, ok := m.trainings[t.ID]; !ok {
		return ErrNotFound
	}
	m.trainings[t.ID] = t
	return nil
}

func (m *mockTrainingRepo) Delete(_ context.Context, id int64) error {
	delete(m.trainings, id)
	return nil
}

var _ repository.TrainingRepository = (*mockTrainingRepo)(nil)

type mockTrainingTrainerRepo struct {
	records map[int64]*entity.TrainingTrainer
	nextID  int64
}

func newMockTrainingTrainerRepo() *mockTrainingTrainerRepo {
	return &mockTrainingTrainerRepo{
		records: make(map[int64]*entity.TrainingTrainer),
		nextID:  1,
	}
}

func (m *mockTrainingTrainerRepo) Create(_ context.Context, tt *entity.TrainingTrainer) (int64, error) {
	tt.ID = m.nextID
	m.records[tt.ID] = tt
	m.nextID++
	return tt.ID, nil
}

func (m *mockTrainingTrainerRepo) ListByTraining(
	_ context.Context,
	trainingID int64,
) ([]entity.TrainingTrainer, error) {
	var result []entity.TrainingTrainer
	for _, tt := range m.records {
		if tt.TrainingID == trainingID {
			result = append(result, *tt)
		}
	}
	return result, nil
}

func (m *mockTrainingTrainerRepo) DeleteByTraining(_ context.Context, trainingID int64) error {
	for id, tt := range m.records {
		if tt.TrainingID == trainingID {
			delete(m.records, id)
		}
	}
	return nil
}

var _ repository.TrainingTrainerRepository = (*mockTrainingTrainerRepo)(nil)

type mockTrainingParticipantRepo struct {
	records map[int64]*entity.TrainingParticipant
	nextID  int64
}

func newMockTrainingParticipantRepo() *mockTrainingParticipantRepo {
	return &mockTrainingParticipantRepo{
		records: make(map[int64]*entity.TrainingParticipant),
		nextID:  1,
	}
}

func (m *mockTrainingParticipantRepo) Create(_ context.Context, tp *entity.TrainingParticipant) (int64, error) {
	tp.ID = m.nextID
	m.records[tp.ID] = tp
	m.nextID++
	return tp.ID, nil
}

func (m *mockTrainingParticipantRepo) GetByTrainingAndMember(
	_ context.Context,
	trainingID, memberID int64,
) (*entity.TrainingParticipant, error) {
	for _, tp := range m.records {
		if tp.TrainingID == trainingID && tp.MemberID == memberID {
			return tp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *mockTrainingParticipantRepo) ListByTraining(
	_ context.Context,
	trainingID int64,
) ([]entity.TrainingParticipant, error) {
	var result []entity.TrainingParticipant
	for _, tp := range m.records {
		if tp.TrainingID == trainingID {
			result = append(result, *tp)
		}
	}
	return result, nil
}

func (m *mockTrainingParticipantRepo) Delete(_ context.Context, id int64) error {
	delete(m.records, id)
	return nil
}

var _ repository.TrainingParticipantRepository = (*mockTrainingParticipantRepo)(nil)

// --- tests ---

func TestCreateTraining(t *testing.T) {
	mr := newMockMemberRepo()
	// Pre-seed trainers
	m1, _ := mr.Create(context.Background(), &entity.Member{FIO: "Trainer1", TelegramID: 1})
	m2, _ := mr.Create(context.Background(), &entity.Member{FIO: "Trainer2", TelegramID: 2})

	tr := newMockTrainingRepo()
	ttr := newMockTrainingTrainerRepo()
	tpr := newMockTrainingParticipantRepo()
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	date := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	training, err := uc.CreateTraining(context.Background(), 1, 5, date, 60, []int64{m1, m2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if training.ID == 0 {
		t.Fatal("expected non-zero training ID")
	}
	if training.Status != entity.TrainingStatusPlanned {
		t.Errorf("expected status %q, got %q", entity.TrainingStatusPlanned, training.Status)
	}
	if training.ClubID != 1 {
		t.Errorf("expected ClubID 1, got %d", training.ClubID)
	}

	// Verify trainer records were created
	trainers, err := ttr.ListByTraining(context.Background(), training.ID)
	if err != nil {
		t.Fatalf("unexpected error listing trainers: %v", err)
	}
	if len(trainers) != 2 {
		t.Errorf("expected 2 trainer records, got %d", len(trainers))
	}

	t.Run("no trainers", func(t *testing.T) {
		noTrainerRun, createErr := uc.CreateTraining(context.Background(), 2, 5, date, 90, nil)
		if createErr != nil {
			t.Fatalf("unexpected error: %v", createErr)
		}
		noTrainers, _ := ttr.ListByTraining(context.Background(), noTrainerRun.ID)
		if len(noTrainers) != 0 {
			t.Errorf("expected 0 trainer records, got %d", len(noTrainers))
		}
	})
}

func TestAddParticipant(t *testing.T) {
	mr := newMockMemberRepo()
	memberID, _ := mr.Create(context.Background(), &entity.Member{FIO: "Runner1", TelegramID: 10})

	tr := newMockTrainingRepo()
	ttr := newMockTrainingTrainerRepo()
	tpr := newMockTrainingParticipantRepo()
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	tid, _ := tr.Create(context.Background(), &entity.Training{ClubID: 1, Status: entity.TrainingStatusPlanned})

	err := uc.AddParticipant(context.Background(), tid, memberID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	participants, _ := tpr.ListByTraining(context.Background(), tid)
	if len(participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(participants))
	}
	if participants[0].MemberID != memberID {
		t.Errorf("expected MemberID %d, got %d", memberID, participants[0].MemberID)
	}
}

func TestRemoveParticipant(t *testing.T) {
	mr := newMockMemberRepo()
	memberID, _ := mr.Create(context.Background(), &entity.Member{FIO: "Runner2", TelegramID: 20})

	tr := newMockTrainingRepo()
	ttr := newMockTrainingTrainerRepo()
	tpr := newMockTrainingParticipantRepo()
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	tid, _ := tr.Create(context.Background(), &entity.Training{ClubID: 1, Status: entity.TrainingStatusPlanned})
	_ = uc.AddParticipant(context.Background(), tid, memberID)

	err := uc.RemoveParticipant(context.Background(), tid, memberID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	participants, _ := tpr.ListByTraining(context.Background(), tid)
	if len(participants) != 0 {
		t.Errorf("expected 0 participants after removal, got %d", len(participants))
	}

	t.Run("remove non-existent participant", func(t *testing.T) {
		remErr := uc.RemoveParticipant(context.Background(), tid, 999)
		if remErr == nil {
			t.Fatal("expected error for non-existent participant")
		}
	})
}

func TestConfirmTraining(t *testing.T) {
	mr := newMockMemberRepo()
	add1, _ := mr.Create(context.Background(), &entity.Member{FIO: "Add1", TelegramID: 30})
	add2, _ := mr.Create(context.Background(), &entity.Member{FIO: "Add2", TelegramID: 31})
	rem1, _ := mr.Create(context.Background(), &entity.Member{FIO: "Rem1", TelegramID: 32})

	tr := newMockTrainingRepo()
	ttr := newMockTrainingTrainerRepo()
	tpr := newMockTrainingParticipantRepo()
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	tid, _ := tr.Create(context.Background(), &entity.Training{ClubID: 1, Status: entity.TrainingStatusConfirming})
	// Pre-add a participant who will be removed
	_, _ = tpr.Create(context.Background(), &entity.TrainingParticipant{TrainingID: tid, MemberID: rem1})

	err := uc.ConfirmTraining(context.Background(), tid, []int64{add1, add2}, []int64{rem1}, "photo123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify status updated to completed
	updated, _ := tr.GetByID(context.Background(), tid)
	if updated.Status != entity.TrainingStatusCompleted {
		t.Errorf("expected status %q, got %q", entity.TrainingStatusCompleted, updated.Status)
	}
	if updated.PhotoFileID != "photo123" {
		t.Errorf("expected PhotoFileID %q, got %q", "photo123", updated.PhotoFileID)
	}

	// Verify added participants
	participants, _ := tpr.ListByTraining(context.Background(), tid)
	memberIDs := make(map[int64]bool)
	for _, p := range participants {
		memberIDs[p.MemberID] = true
	}
	if !memberIDs[add1] || !memberIDs[add2] {
		t.Error("expected add1 and add2 to be in participants")
	}
	if memberIDs[rem1] {
		t.Error("expected rem1 to be removed from participants")
	}
}

func TestFindTrainingsNeedingConfirmation(t *testing.T) {
	mr := newMockMemberRepo()
	tr := newMockTrainingRepo()
	ttr := newMockTrainingTrainerRepo()
	tpr := newMockTrainingParticipantRepo()
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	now := time.Now()

	// Training that ended more than 1 hour ago (should be included)
	past := now.Add(-3 * time.Hour)
	_, _ = tr.Create(context.Background(), &entity.Training{
		ClubID:   1,
		Date:     past,
		Duration: 60,
		Status:   entity.TrainingStatusPlanned,
	})

	// Training that ended less than 1 hour ago (should NOT be included)
	recent := now.Add(-30 * time.Minute)
	_, _ = tr.Create(context.Background(), &entity.Training{
		ClubID:   2,
		Date:     recent,
		Duration: 60,
		Status:   entity.TrainingStatusPlanned,
	})

	// Completed training (should NOT be included regardless of time)
	_, _ = tr.Create(context.Background(), &entity.Training{
		ClubID:   3,
		Date:     past,
		Duration: 60,
		Status:   entity.TrainingStatusCompleted,
	})

	// In-progress training that ended more than 1 hour ago (should be included)
	_, _ = tr.Create(context.Background(), &entity.Training{
		ClubID:   4,
		Date:     past,
		Duration: 60,
		Status:   entity.TrainingStatusInProgress,
	})

	result, err := uc.FindTrainingsNeedingConfirmation(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 trainings needing confirmation, got %d", len(result))
	}

	// Verify the right ones were returned
	for _, training := range result {
		if training.Status != entity.TrainingStatusPlanned && training.Status != entity.TrainingStatusInProgress {
			t.Errorf("unexpected status %q in result", training.Status)
		}
	}
}
