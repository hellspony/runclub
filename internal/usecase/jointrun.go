package usecase

import (
	"context"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

type JointRunUseCase interface {
	CreateJointRun(ctx context.Context, clubID, locationID, creatorID int64, date time.Time) (*entity.JointRun, error)
	GetJointRun(ctx context.Context, id int64) (*entity.JointRun, error)
	ListJointRuns(ctx context.Context, clubID int64) ([]entity.JointRun, error)
	UpdateJointRun(ctx context.Context, run *entity.JointRun) error
	DeleteJointRun(ctx context.Context, id int64) error
	AddParticipant(ctx context.Context, runID, memberID int64) error
	RemoveParticipant(ctx context.Context, runID, memberID int64) error
	ListParticipants(ctx context.Context, runID int64) ([]entity.Member, error)
}

type jointRunUseCase struct {
	jointRunRepo     repository.JointRunRepository
	jointRunPartRepo repository.JointRunParticipantRepository
	memberRepo       repository.MemberRepository
}

func NewJointRunUseCase(
	jointRunRepo repository.JointRunRepository,
	jointRunPartRepo repository.JointRunParticipantRepository,
	memberRepo repository.MemberRepository,
) JointRunUseCase {
	return &jointRunUseCase{
		jointRunRepo:     jointRunRepo,
		jointRunPartRepo: jointRunPartRepo,
		memberRepo:       memberRepo,
	}
}

func (uc *jointRunUseCase) CreateJointRun(
	ctx context.Context,
	clubID, locationID, creatorID int64,
	date time.Time,
) (*entity.JointRun, error) {
	run := &entity.JointRun{
		ClubID:     clubID,
		LocationID: locationID,
		CreatorID:  creatorID,
		Date:       date,
	}

	id, err := uc.jointRunRepo.Create(ctx, run)
	if err != nil {
		return nil, err
	}
	run.ID = id

	return run, nil
}

func (uc *jointRunUseCase) GetJointRun(ctx context.Context, id int64) (*entity.JointRun, error) {
	return uc.jointRunRepo.GetByID(ctx, id)
}

func (uc *jointRunUseCase) ListJointRuns(ctx context.Context, clubID int64) ([]entity.JointRun, error) {
	return uc.jointRunRepo.ListByClub(ctx, clubID)
}

func (uc *jointRunUseCase) UpdateJointRun(ctx context.Context, run *entity.JointRun) error {
	return uc.jointRunRepo.Update(ctx, run)
}

func (uc *jointRunUseCase) DeleteJointRun(ctx context.Context, id int64) error {
	return uc.jointRunRepo.Delete(ctx, id)
}

func (uc *jointRunUseCase) AddParticipant(ctx context.Context, runID, memberID int64) error {
	p := &entity.JointRunParticipant{
		JointRunID: runID,
		MemberID:   memberID,
	}

	_, err := uc.jointRunPartRepo.Create(ctx, p)
	return err
}

func (uc *jointRunUseCase) RemoveParticipant(ctx context.Context, runID, memberID int64) error {
	p, err := uc.jointRunPartRepo.GetByRunAndMember(ctx, runID, memberID)
	if err != nil {
		return err
	}

	return uc.jointRunPartRepo.Delete(ctx, p.ID)
}

func (uc *jointRunUseCase) ListParticipants(ctx context.Context, runID int64) ([]entity.Member, error) {
	participants, err := uc.jointRunPartRepo.ListByRun(ctx, runID)
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
