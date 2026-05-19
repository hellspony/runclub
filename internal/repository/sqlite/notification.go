package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RaceNotificationLogRepositoryImpl struct {
	db *sqlx.DB
}

func NewRaceNotificationLogRepository(db *sqlx.DB) *RaceNotificationLogRepositoryImpl {
	return &RaceNotificationLogRepositoryImpl{db: db}
}

func (r *RaceNotificationLogRepositoryImpl) Create(ctx context.Context, clubID, raceID int64, sentDate string) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO race_notification_log (club_id, race_id, sent_date)
		 VALUES (:club_id, :race_id, :sent_date)`,
		map[string]any{
			"club_id":   clubID,
			"race_id":   raceID,
			"sent_date": sentDate,
		},
	)
	if err != nil {
		return fmt.Errorf("insert race notification log: %w", err)
	}
	return nil
}

func (r *RaceNotificationLogRepositoryImpl) Exists(
	ctx context.Context,
	clubID, raceID int64,
	sentDate string,
) (bool, error) {
	var id int64
	err := r.db.GetContext(ctx, &id,
		`SELECT id FROM race_notification_log WHERE club_id = ? AND race_id = ? AND sent_date = ?`,
		clubID, raceID, sentDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check race notification log exists: %w", err)
	}
	return true, nil
}
