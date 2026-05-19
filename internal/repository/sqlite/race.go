package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"runclub/internal/domain/entity"
)

type RaceRepositoryImpl struct {
	db *sqlx.DB
}

func NewRaceRepository(db *sqlx.DB) *RaceRepositoryImpl {
	return &RaceRepositoryImpl{db: db}
}

type raceRow struct {
	ID        int64  `db:"id"`
	ClubID    int64  `db:"club_id"`
	Date      string `db:"date"`
	Type      string `db:"type"`
	Place     string `db:"place"`
	Distances string `db:"distances"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type RaceRegistrationRepositoryImpl struct {
	db *sqlx.DB
}

func NewRaceRegistrationRepository(db *sqlx.DB) *RaceRegistrationRepositoryImpl {
	return &RaceRegistrationRepositoryImpl{db: db}
}

type raceRegistrationRow struct {
	ID        int64  `db:"id"`
	RaceID    int64  `db:"race_id"`
	MemberID  int64  `db:"member_id"`
	Distance  string `db:"distance"`
	CreatedAt string `db:"created_at"`
}

func (r *RaceRepositoryImpl) Create(ctx context.Context, race *entity.Race) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO races (club_id, date, type, place, distances, name, created_at, updated_at)
		 VALUES (:club_id, :date, :type, :place, :distances, :name, :created_at, :updated_at)`,
		map[string]any{
			"club_id":    race.ClubID,
			"date":       race.Date.Format(time.RFC3339),
			"type":       race.Type,
			"place":      race.Place,
			"distances":  race.Distances,
			"name":       race.Name,
			"created_at": now,
			"updated_at": now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert race: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *RaceRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Race, error) {
	var row raceRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, club_id, date, type, place, distances, name, created_at, updated_at FROM races WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("race not found: %w", err)
		}
		return nil, fmt.Errorf("get race by id: %w", err)
	}
	return raceRowToEntity(&row), nil
}

func (r *RaceRepositoryImpl) ListByClub(ctx context.Context, clubID int64) ([]entity.Race, error) {
	var rows []raceRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, club_id, date, type, place, distances, name, created_at, updated_at FROM races WHERE club_id = ? ORDER BY date`,
		clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("list races by club: %w", err)
	}
	result := make([]entity.Race, len(rows))
	for i := range rows {
		result[i] = *raceRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *RaceRepositoryImpl) ListUpcomingByClub(
	ctx context.Context,
	clubID int64,
	from, to time.Time,
) ([]entity.Race, error) {
	var rows []raceRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, club_id, date, type, place, distances, name, created_at, updated_at FROM races WHERE club_id = ? AND date >= ? AND date <= ? ORDER BY date`,
		clubID,
		from.Format(time.RFC3339),
		to.Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("list upcoming races by club: %w", err)
	}
	result := make([]entity.Race, len(rows))
	for i := range rows {
		result[i] = *raceRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *RaceRepositoryImpl) Update(ctx context.Context, race *entity.Race) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE races SET club_id = :club_id, date = :date, type = :type,
		 place = :place, distances = :distances, name = :name, updated_at = :updated_at
		 WHERE id = :id`,
		map[string]any{
			"id":         race.ID,
			"club_id":    race.ClubID,
			"date":       race.Date.Format(time.RFC3339),
			"type":       race.Type,
			"place":      race.Place,
			"distances":  race.Distances,
			"name":       race.Name,
			"updated_at": now,
		},
	)
	if err != nil {
		return fmt.Errorf("update race: %w", err)
	}
	return nil
}

func (r *RaceRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM races WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete race: %w", err)
	}
	return nil
}

func raceRowToEntity(row *raceRow) *entity.Race {
	date, _ := time.Parse(time.RFC3339, row.Date)
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)
	return &entity.Race{
		ID:        row.ID,
		ClubID:    row.ClubID,
		Date:      date,
		Type:      row.Type,
		Place:     row.Place,
		Distances: row.Distances,
		Name:      row.Name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// RaceRegistrationRepositoryImpl methods

func (r *RaceRegistrationRepositoryImpl) Create(ctx context.Context, reg *entity.RaceRegistration) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO race_registrations (race_id, member_id, distance, created_at)
		 VALUES (:race_id, :member_id, :distance, :created_at)`,
		map[string]any{
			"race_id":    reg.RaceID,
			"member_id":  reg.MemberID,
			"distance":   reg.Distance,
			"created_at": now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert race registration: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *RaceRegistrationRepositoryImpl) GetByRaceAndMember(
	ctx context.Context,
	raceID, memberID int64,
) (*entity.RaceRegistration, error) {
	var row raceRegistrationRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, race_id, member_id, distance, created_at FROM race_registrations WHERE race_id = ? AND member_id = ?`,
		raceID,
		memberID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("race registration not found: %w", err)
		}
		return nil, fmt.Errorf("get race registration: %w", err)
	}
	return raceRegistrationRowToEntity(&row), nil
}

func (r *RaceRegistrationRepositoryImpl) ListByRace(
	ctx context.Context,
	raceID int64,
) ([]entity.RaceRegistration, error) {
	var rows []raceRegistrationRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, race_id, member_id, distance, created_at FROM race_registrations WHERE race_id = ?`, raceID)
	if err != nil {
		return nil, fmt.Errorf("list race registrations by race: %w", err)
	}
	result := make([]entity.RaceRegistration, len(rows))
	for i := range rows {
		result[i] = *raceRegistrationRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *RaceRegistrationRepositoryImpl) ListByMember(
	ctx context.Context,
	memberID int64,
) ([]entity.RaceRegistration, error) {
	var rows []raceRegistrationRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, race_id, member_id, distance, created_at FROM race_registrations WHERE member_id = ?`, memberID)
	if err != nil {
		return nil, fmt.Errorf("list race registrations by member: %w", err)
	}
	result := make([]entity.RaceRegistration, len(rows))
	for i := range rows {
		result[i] = *raceRegistrationRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *RaceRegistrationRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM race_registrations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete race registration: %w", err)
	}
	return nil
}

func raceRegistrationRowToEntity(row *raceRegistrationRow) *entity.RaceRegistration {
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	return &entity.RaceRegistration{
		ID:        row.ID,
		RaceID:    row.RaceID,
		MemberID:  row.MemberID,
		Distance:  row.Distance,
		CreatedAt: createdAt,
	}
}
