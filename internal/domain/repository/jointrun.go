package repository

import (
	"context"

	"runclub/internal/domain/entity"
)

type JointRunRepository interface {
	Create(ctx context.Context, run *entity.JointRun) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.JointRun, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.JointRun, error)
	Update(ctx context.Context, run *entity.JointRun) error
	Delete(ctx context.Context, id int64) error
}

type JointRunParticipantRepository interface {
	Create(ctx context.Context, p *entity.JointRunParticipant) (int64, error)
	GetByRunAndMember(ctx context.Context, runID, memberID int64) (*entity.JointRunParticipant, error)
	ListByRun(ctx context.Context, runID int64) ([]entity.JointRunParticipant, error)
	Delete(ctx context.Context, id int64) error
}
