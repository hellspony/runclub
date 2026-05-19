package repository

import (
	"context"
)

type RaceNotificationLogRepository interface {
	Create(ctx context.Context, clubID, raceID int64, sentDate string) error
	Exists(ctx context.Context, clubID, raceID int64, sentDate string) (bool, error)
}
