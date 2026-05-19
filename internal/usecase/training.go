package usecase

import (
	"context"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

type TrainingUseCase interface {
	CreateTraining(
		ctx context.Context,
		clubID, locationID int64,
		date time.Time,
		duration int,
		trainerIDs []int64,
	) (*entity.Training, error)
	GetTraining(ctx context.Context, id int64) (*entity.Training, error)
	ListTrainings(ctx context.Context, clubID int64) ([]entity.Training, error)
	UpdateTraining(ctx context.Context, training *entity.Training) error
	DeleteTraining(ctx context.Context, id int64) error
	AddParticipant(ctx context.Context, trainingID, memberID int64) error
	RemoveParticipant(ctx context.Context, trainingID, memberID int64) error
	ListParticipants(ctx context.Context, trainingID int64) ([]entity.Member, error)
	ListTrainers(ctx context.Context, trainingID int64) ([]entity.Member, error)
	StartConfirmation(ctx context.Context, trainingID int64) error
	ConfirmTraining(ctx context.Context, trainingID int64, addedIDs, removedIDs []int64, photoFileID string) error
	FindTrainingsNeedingConfirmation(ctx context.Context) ([]entity.Training, error)
}

type trainingUseCase struct {
	trainingRepo    repository.TrainingRepository
	trainerRepo     repository.TrainingTrainerRepository
	participantRepo repository.TrainingParticipantRepository
	memberRepo      repository.MemberRepository
}

func NewTrainingUseCase(
	trainingRepo repository.TrainingRepository,
	trainerRepo repository.TrainingTrainerRepository,
	participantRepo repository.TrainingParticipantRepository,
	memberRepo repository.MemberRepository,
) TrainingUseCase {
	return &trainingUseCase{
		trainingRepo:    trainingRepo,
		trainerRepo:     trainerRepo,
		participantRepo: participantRepo,
		memberRepo:      memberRepo,
	}
}

func (uc *trainingUseCase) CreateTraining(
	ctx context.Context,
	clubID, locationID int64,
	date time.Time,
	duration int,
	trainerIDs []int64,
) (*entity.Training, error) {
	training := &entity.Training{
		ClubID:     clubID,
		LocationID: locationID,
		Date:       date,
		Duration:   duration,
		Status:     entity.TrainingStatusPlanned,
	}

	id, err := uc.trainingRepo.Create(ctx, training)
	if err != nil {
		return nil, err
	}
	training.ID = id

	for _, trainerID := range trainerIDs {
		tt := &entity.TrainingTrainer{
			TrainingID: id,
			MemberID:   trainerID,
		}
		if _, trainerErr := uc.trainerRepo.Create(ctx, tt); trainerErr != nil {
			return nil, trainerErr
		}
	}

	return training, nil
}

func (uc *trainingUseCase) GetTraining(ctx context.Context, id int64) (*entity.Training, error) {
	return uc.trainingRepo.GetByID(ctx, id)
}

func (uc *trainingUseCase) ListTrainings(ctx context.Context, clubID int64) ([]entity.Training, error) {
	return uc.trainingRepo.ListByClub(ctx, clubID)
}

func (uc *trainingUseCase) UpdateTraining(ctx context.Context, training *entity.Training) error {
	return uc.trainingRepo.Update(ctx, training)
}

func (uc *trainingUseCase) DeleteTraining(ctx context.Context, id int64) error {
	return uc.trainingRepo.Delete(ctx, id)
}

func (uc *trainingUseCase) AddParticipant(ctx context.Context, trainingID, memberID int64) error {
	tp := &entity.TrainingParticipant{
		TrainingID: trainingID,
		MemberID:   memberID,
	}

	_, err := uc.participantRepo.Create(ctx, tp)
	return err
}

func (uc *trainingUseCase) RemoveParticipant(ctx context.Context, trainingID, memberID int64) error {
	tp, err := uc.participantRepo.GetByTrainingAndMember(ctx, trainingID, memberID)
	if err != nil {
		return err
	}

	return uc.participantRepo.Delete(ctx, tp.ID)
}

func (uc *trainingUseCase) ListParticipants(ctx context.Context, trainingID int64) ([]entity.Member, error) {
	participants, err := uc.participantRepo.ListByTraining(ctx, trainingID)
	if err != nil {
		return nil, err
	}

	members := make([]entity.Member, 0, len(participants))
	for _, p := range participants {
		member, memberErr := uc.memberRepo.GetByID(ctx, p.MemberID)
		if memberErr != nil {
			return nil, memberErr
		}
		members = append(members, *member)
	}

	return members, nil
}

func (uc *trainingUseCase) ListTrainers(ctx context.Context, trainingID int64) ([]entity.Member, error) {
	trainers, err := uc.trainerRepo.ListByTraining(ctx, trainingID)
	if err != nil {
		return nil, err
	}

	members := make([]entity.Member, 0, len(trainers))
	for _, t := range trainers {
		member, memberErr := uc.memberRepo.GetByID(ctx, t.MemberID)
		if memberErr != nil {
			return nil, memberErr
		}
		members = append(members, *member)
	}

	return members, nil
}

func (uc *trainingUseCase) StartConfirmation(ctx context.Context, trainingID int64) error {
	training, err := uc.trainingRepo.GetByID(ctx, trainingID)
	if err != nil {
		return err
	}

	training.Status = entity.TrainingStatusConfirming
	return uc.trainingRepo.Update(ctx, training)
}

func (uc *trainingUseCase) ConfirmTraining(
	ctx context.Context,
	trainingID int64,
	addedIDs, removedIDs []int64,
	photoFileID string,
) error {
	for _, memberID := range addedIDs {
		tp := &entity.TrainingParticipant{
			TrainingID: trainingID,
			MemberID:   memberID,
		}
		if _, err := uc.participantRepo.Create(ctx, tp); err != nil {
			return err
		}
	}

	for _, memberID := range removedIDs {
		tp, err := uc.participantRepo.GetByTrainingAndMember(ctx, trainingID, memberID)
		if err != nil {
			return err
		}
		if delErr := uc.participantRepo.Delete(ctx, tp.ID); delErr != nil {
			return delErr
		}
	}

	training, err := uc.trainingRepo.GetByID(ctx, trainingID)
	if err != nil {
		return err
	}

	training.Status = entity.TrainingStatusCompleted
	training.PhotoFileID = photoFileID

	return uc.trainingRepo.Update(ctx, training)
}

func (uc *trainingUseCase) FindTrainingsNeedingConfirmation(ctx context.Context) ([]entity.Training, error) {
	planned, err := uc.trainingRepo.ListByStatus(ctx, entity.TrainingStatusPlanned)
	if err != nil {
		return nil, err
	}

	inProgress, err := uc.trainingRepo.ListByStatus(ctx, entity.TrainingStatusInProgress)
	if err != nil {
		return nil, err
	}

	all := make([]entity.Training, 0, len(planned)+len(inProgress))
	all = append(all, planned...)
	all = append(all, inProgress...)

	now := time.Now()
	var result []entity.Training
	for _, t := range all {
		endTime := t.Date.Add(time.Duration(t.Duration) * time.Minute)
		threshold := endTime.Add(1 * time.Hour)
		if threshold.Before(now) {
			result = append(result, t)
		}
	}

	return result, nil
}
