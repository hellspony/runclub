package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	go_mock "go.uber.org/mock/gomock"

	"runclub/internal/domain/entity"
	"runclub/internal/mocks"
	"runclub/internal/usecase"
)

func TestCreateTraining(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	tr := mocks.NewMockTrainingRepository(ctrl)
	ttr := mocks.NewMockTrainingTrainerRepository(ctrl)
	tpr := mocks.NewMockTrainingParticipantRepository(ctrl)
	mr := mocks.NewMockMemberRepository(ctrl)
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	date := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)

	t.Run("with trainers", func(t *testing.T) {
		tr.EXPECT().
			Create(go_mock.Any(), go_mock.Any()).
			Return(int64(1), nil)
		ttr.EXPECT().Create(go_mock.Any(), go_mock.Any()).Return(int64(1), nil)
		ttr.EXPECT().Create(go_mock.Any(), go_mock.Any()).Return(int64(2), nil)

		training, err := uc.CreateTraining(context.Background(), 1, 5, date, 60, []int64{10, 20})
		require.NoError(t, err)
		assert.Equal(t, int64(1), training.ID)
		assert.Equal(t, entity.TrainingStatusPlanned, training.Status)
		assert.Equal(t, int64(1), training.ClubID)
	})

	t.Run("no trainers", func(t *testing.T) {
		tr.EXPECT().
			Create(go_mock.Any(), go_mock.Any()).
			Return(int64(2), nil)

		noTrainerRun, createErr := uc.CreateTraining(context.Background(), 2, 5, date, 90, nil)
		require.NoError(t, createErr)
		assert.Equal(t, int64(2), noTrainerRun.ID)
	})
}

func TestAddParticipant(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	tr := mocks.NewMockTrainingRepository(ctrl)
	ttr := mocks.NewMockTrainingTrainerRepository(ctrl)
	tpr := mocks.NewMockTrainingParticipantRepository(ctrl)
	mr := mocks.NewMockMemberRepository(ctrl)
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	tpr.EXPECT().
		Create(go_mock.Any(), go_mock.Any()).
		Return(int64(1), nil)

	err := uc.AddParticipant(context.Background(), 1, 10)
	require.NoError(t, err)
}

func TestRemoveParticipant(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	tr := mocks.NewMockTrainingRepository(ctrl)
	ttr := mocks.NewMockTrainingTrainerRepository(ctrl)
	tpr := mocks.NewMockTrainingParticipantRepository(ctrl)
	mr := mocks.NewMockMemberRepository(ctrl)
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	tpr.EXPECT().
		GetByTrainingAndMember(go_mock.Any(), int64(1), int64(10)).
		Return(&entity.TrainingParticipant{ID: 5, TrainingID: 1, MemberID: 10}, nil)
	tpr.EXPECT().
		Delete(go_mock.Any(), int64(5)).
		Return(nil)

	err := uc.RemoveParticipant(context.Background(), 1, 10)
	require.NoError(t, err)

	t.Run("remove non-existent participant", func(t *testing.T) {
		tpr.EXPECT().
			GetByTrainingAndMember(go_mock.Any(), int64(1), int64(999)).
			Return(nil, assert.AnError)

		remErr := uc.RemoveParticipant(context.Background(), 1, 999)
		assert.Error(t, remErr)
	})
}

func TestConfirmTraining(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	tr := mocks.NewMockTrainingRepository(ctrl)
	ttr := mocks.NewMockTrainingTrainerRepository(ctrl)
	tpr := mocks.NewMockTrainingParticipantRepository(ctrl)
	mr := mocks.NewMockMemberRepository(ctrl)
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	// Add participants
	tpr.EXPECT().Create(go_mock.Any(), go_mock.Any()).Return(int64(1), nil)
	tpr.EXPECT().Create(go_mock.Any(), go_mock.Any()).Return(int64(2), nil)

	// Remove participant
	tpr.EXPECT().
		GetByTrainingAndMember(go_mock.Any(), int64(1), int64(32)).
		Return(&entity.TrainingParticipant{ID: 100, TrainingID: 1, MemberID: 32}, nil)
	tpr.EXPECT().Delete(go_mock.Any(), int64(100)).Return(nil)

	// Update training status
	tr.EXPECT().
		GetByID(go_mock.Any(), int64(1)).
		Return(&entity.Training{
			ID:     1,
			ClubID: 1,
			Status: entity.TrainingStatusConfirming,
		}, nil)
	tr.EXPECT().
		Update(go_mock.Any(), go_mock.Any()).
		Return(nil)

	err := uc.ConfirmTraining(context.Background(), 1, []int64{30, 31}, []int64{32}, "photo123")
	require.NoError(t, err)
}

func TestFindTrainingsNeedingConfirmation(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	tr := mocks.NewMockTrainingRepository(ctrl)
	ttr := mocks.NewMockTrainingTrainerRepository(ctrl)
	tpr := mocks.NewMockTrainingParticipantRepository(ctrl)
	mr := mocks.NewMockMemberRepository(ctrl)
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	now := time.Now()
	past := now.Add(-3 * time.Hour)
	recent := now.Add(-30 * time.Minute)

	// Planned trainings
	tr.EXPECT().
		ListByStatus(go_mock.Any(), entity.TrainingStatusPlanned).
		Return([]entity.Training{
			{ID: 1, ClubID: 1, Date: past, Duration: 60, Status: entity.TrainingStatusPlanned},
			{ID: 2, ClubID: 2, Date: recent, Duration: 60, Status: entity.TrainingStatusPlanned},
		}, nil)

	// In-progress trainings
	tr.EXPECT().
		ListByStatus(go_mock.Any(), entity.TrainingStatusInProgress).
		Return([]entity.Training{
			{ID: 3, ClubID: 3, Date: past, Duration: 60, Status: entity.TrainingStatusInProgress},
		}, nil)

	result, err := uc.FindTrainingsNeedingConfirmation(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)

	for _, training := range result {
		assert.Contains(t, []entity.TrainingStatus{
			entity.TrainingStatusPlanned,
			entity.TrainingStatusInProgress,
		}, training.Status)
	}
}

func TestStartConfirmation(t *testing.T) {
	ctrl := go_mock.NewController(t)
	defer ctrl.Finish()

	tr := mocks.NewMockTrainingRepository(ctrl)
	ttr := mocks.NewMockTrainingTrainerRepository(ctrl)
	tpr := mocks.NewMockTrainingParticipantRepository(ctrl)
	mr := mocks.NewMockMemberRepository(ctrl)
	uc := usecase.NewTrainingUseCase(tr, ttr, tpr, mr)

	tr.EXPECT().
		GetByID(go_mock.Any(), int64(1)).
		Return(&entity.Training{ID: 1, Status: entity.TrainingStatusPlanned}, nil)
	tr.EXPECT().
		Update(go_mock.Any(), go_mock.Any()).
		Return(nil)

	err := uc.StartConfirmation(context.Background(), 1)
	require.NoError(t, err)
}
