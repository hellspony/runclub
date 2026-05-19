package usecase

import (
	"context"
	"time"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

type MemberUseCase interface {
	RegisterOrGet(ctx context.Context, telegramID int64, username string) (*entity.Member, error)
	CreateMember(ctx context.Context, member *entity.Member) (int64, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.Member, error)
	GetByID(ctx context.Context, id int64) (*entity.Member, error)
	UpdateProfile(ctx context.Context, id int64, fio string, telegramUsername string, birthDate *time.Time) error
	AddToClub(ctx context.Context, clubID, memberID int64, role entity.MemberRole) error
	RemoveFromClub(ctx context.Context, clubID, memberID int64) error
	GetClubMember(ctx context.Context, clubID, memberID int64) (*entity.ClubMember, error)
	ListMembers(ctx context.Context, clubID int64) ([]entity.Member, error)
	ListTrainers(ctx context.Context, clubID int64) ([]entity.Member, error)
	ListTrainerClubs(ctx context.Context, memberID int64) ([]entity.ClubMember, error)
	ListClubsByMember(ctx context.Context, memberID int64) ([]entity.ClubMember, error)
	HasClub(ctx context.Context, memberID int64) (bool, error)
	CleanupOrphanMembers(ctx context.Context, olderThanDays int) (int, error)
	DeleteMember(ctx context.Context, id int64) error
	UpdateRole(ctx context.Context, clubID, memberID int64, role entity.MemberRole) error
}

type memberUseCase struct {
	memberRepo     repository.MemberRepository
	clubMemberRepo repository.ClubMemberRepository
}

func NewMemberUseCase(
	memberRepo repository.MemberRepository,
	clubMemberRepo repository.ClubMemberRepository,
) MemberUseCase {
	return &memberUseCase{
		memberRepo:     memberRepo,
		clubMemberRepo: clubMemberRepo,
	}
}

func (uc *memberUseCase) RegisterOrGet(ctx context.Context, telegramID int64, username string) (*entity.Member, error) {
	member, err := uc.memberRepo.GetByTelegramID(ctx, telegramID)
	if err == nil && member != nil {
		return member, nil
	}

	member = &entity.Member{
		FIO:              username,
		TelegramUsername: username,
		TelegramID:       telegramID,
	}

	id, err := uc.memberRepo.Create(ctx, member)
	if err != nil {
		return nil, err
	}

	member.ID = id
	return member, nil
}

func (uc *memberUseCase) CreateMember(ctx context.Context, member *entity.Member) (int64, error) {
	return uc.memberRepo.Create(ctx, member)
}

func (uc *memberUseCase) GetByTelegramID(ctx context.Context, telegramID int64) (*entity.Member, error) {
	return uc.memberRepo.GetByTelegramID(ctx, telegramID)
}

func (uc *memberUseCase) GetByID(ctx context.Context, id int64) (*entity.Member, error) {
	return uc.memberRepo.GetByID(ctx, id)
}

func (uc *memberUseCase) UpdateProfile(
	ctx context.Context,
	id int64,
	fio string,
	telegramUsername string,
	birthDate *time.Time,
) error {
	member, err := uc.memberRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	member.FIO = fio
	member.TelegramUsername = telegramUsername
	member.BirthDate = birthDate

	return uc.memberRepo.Update(ctx, member)
}

func (uc *memberUseCase) AddToClub(ctx context.Context, clubID, memberID int64, role entity.MemberRole) error {
	cm := &entity.ClubMember{
		ClubID:   clubID,
		MemberID: memberID,
		Role:     role,
	}

	_, err := uc.clubMemberRepo.Create(ctx, cm)
	if err != nil {
		return err
	}

	// Clear left_at since member is back in a club
	member, err := uc.memberRepo.GetByID(ctx, memberID)
	if err == nil && member.LeftAt != nil {
		member.LeftAt = nil
		_ = uc.memberRepo.Update(ctx, member)
	}

	return nil
}

func (uc *memberUseCase) RemoveFromClub(ctx context.Context, clubID, memberID int64) error {
	if err := uc.clubMemberRepo.Delete(ctx, clubID, memberID); err != nil {
		return err
	}

	// If member has no clubs left, set left_at
	hasClub, err := uc.HasClub(ctx, memberID)
	if err == nil && !hasClub {
		var member *entity.Member
		member, err = uc.memberRepo.GetByID(ctx, memberID)
		if err == nil {
			now := time.Now()
			member.LeftAt = &now
			_ = uc.memberRepo.Update(ctx, member)
		}
	}

	return nil
}

func (uc *memberUseCase) GetClubMember(ctx context.Context, clubID, memberID int64) (*entity.ClubMember, error) {
	return uc.clubMemberRepo.GetByClubAndMember(ctx, clubID, memberID)
}

func (uc *memberUseCase) ListMembers(ctx context.Context, clubID int64) ([]entity.Member, error) {
	return uc.memberRepo.ListByClub(ctx, clubID)
}

func (uc *memberUseCase) ListTrainers(ctx context.Context, clubID int64) ([]entity.Member, error) {
	return uc.memberRepo.ListTrainersByClub(ctx, clubID)
}

func (uc *memberUseCase) ListTrainerClubs(ctx context.Context, memberID int64) ([]entity.ClubMember, error) {
	return uc.clubMemberRepo.ListTrainerClubs(ctx, memberID)
}

func (uc *memberUseCase) ListClubsByMember(ctx context.Context, memberID int64) ([]entity.ClubMember, error) {
	return uc.clubMemberRepo.ListClubsByMember(ctx, memberID)
}

func (uc *memberUseCase) HasClub(ctx context.Context, memberID int64) (bool, error) {
	clubs, err := uc.clubMemberRepo.ListClubsByMember(ctx, memberID)
	if err != nil {
		return false, err
	}
	return len(clubs) > 0, nil
}

func (uc *memberUseCase) CleanupOrphanMembers(ctx context.Context, olderThanDays int) (int, error) {
	orphans, err := uc.memberRepo.ListOrphansOlderThan(ctx, olderThanDays)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, m := range orphans {
		if err = uc.memberRepo.Delete(ctx, m.ID); err != nil {
			continue
		}
		deleted++
	}
	return deleted, nil
}

func (uc *memberUseCase) DeleteMember(ctx context.Context, id int64) error {
	return uc.memberRepo.Delete(ctx, id)
}

func (uc *memberUseCase) UpdateRole(ctx context.Context, clubID, memberID int64, role entity.MemberRole) error {
	return uc.clubMemberRepo.UpdateRole(ctx, clubID, memberID, role)
}
