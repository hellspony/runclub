package repository

import (
	"context"
	"time"

	"runclub/internal/domain/entity"
)

type RaceRepository interface {
	Create(ctx context.Context, race *entity.Race) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Race, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.Race, error)
	ListUpcomingByClub(ctx context.Context, clubID int64, from, to time.Time) ([]entity.Race, error)
	Update(ctx context.Context, race *entity.Race) error
	Delete(ctx context.Context, id int64) error
}

type RaceRegistrationRepository interface {
	Create(ctx context.Context, reg *entity.RaceRegistration) (int64, error)
	GetByRaceAndMember(ctx context.Context, raceID, memberID int64) (*entity.RaceRegistration, error)
	ListByRace(ctx context.Context, raceID int64) ([]entity.RaceRegistration, error)
	ListByMember(ctx context.Context, memberID int64) ([]entity.RaceRegistration, error)
	Delete(ctx context.Context, id int64) error
}
