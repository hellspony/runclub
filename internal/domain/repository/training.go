package repository

import (
	"context"

	"runclub/internal/domain/entity"
)

type TrainingRepository interface {
	Create(ctx context.Context, training *entity.Training) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Training, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.Training, error)
	ListByStatus(ctx context.Context, status entity.TrainingStatus) ([]entity.Training, error)
	Update(ctx context.Context, training *entity.Training) error
	Delete(ctx context.Context, id int64) error
}

type TrainingTrainerRepository interface {
	Create(ctx context.Context, tt *entity.TrainingTrainer) (int64, error)
	ListByTraining(ctx context.Context, trainingID int64) ([]entity.TrainingTrainer, error)
	DeleteByTraining(ctx context.Context, trainingID int64) error
}

type TrainingParticipantRepository interface {
	Create(ctx context.Context, tp *entity.TrainingParticipant) (int64, error)
	GetByTrainingAndMember(ctx context.Context, trainingID, memberID int64) (*entity.TrainingParticipant, error)
	ListByTraining(ctx context.Context, trainingID int64) ([]entity.TrainingParticipant, error)
	Delete(ctx context.Context, id int64) error
}
