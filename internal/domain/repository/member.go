package repository

import (
	"context"

	"runclub/internal/domain/entity"
)

type MemberRepository interface {
	Create(ctx context.Context, member *entity.Member) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Member, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.Member, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.Member, error)
	ListTrainersByClub(ctx context.Context, clubID int64) ([]entity.Member, error)
	ListBirthdayOn(ctx context.Context, month, day int) ([]entity.Member, error)
	ListOrphansOlderThan(ctx context.Context, days int) ([]entity.Member, error)
	Update(ctx context.Context, member *entity.Member) error
	Delete(ctx context.Context, id int64) error
}

type ClubMemberRepository interface {
	Create(ctx context.Context, cm *entity.ClubMember) (int64, error)
	GetByClubAndMember(ctx context.Context, clubID, memberID int64) (*entity.ClubMember, error)
	ListClubsByMember(ctx context.Context, memberID int64) ([]entity.ClubMember, error)
	ListTrainerClubs(ctx context.Context, memberID int64) ([]entity.ClubMember, error)
	UpdateRole(ctx context.Context, clubID, memberID int64, role entity.MemberRole) error
	Delete(ctx context.Context, clubID, memberID int64) error
}
