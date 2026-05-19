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

type TrainingRepositoryImpl struct {
	db *sqlx.DB
}

func NewTrainingRepository(db *sqlx.DB) *TrainingRepositoryImpl {
	return &TrainingRepositoryImpl{db: db}
}

type trainingRow struct {
	ID          int64  `db:"id"`
	ClubID      int64  `db:"club_id"`
	LocationID  int64  `db:"location_id"`
	Date        string `db:"date"`
	Duration    int    `db:"duration"`
	Status      string `db:"status"`
	PhotoFileID string `db:"photo_file_id"`
	MessageID   int64  `db:"message_id"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

type TrainingTrainerRepositoryImpl struct {
	db *sqlx.DB
}

func NewTrainingTrainerRepository(db *sqlx.DB) *TrainingTrainerRepositoryImpl {
	return &TrainingTrainerRepositoryImpl{db: db}
}

type trainingTrainerRow struct {
	ID         int64 `db:"id"`
	TrainingID int64 `db:"training_id"`
	MemberID   int64 `db:"member_id"`
}

type TrainingParticipantRepositoryImpl struct {
	db *sqlx.DB
}

func NewTrainingParticipantRepository(db *sqlx.DB) *TrainingParticipantRepositoryImpl {
	return &TrainingParticipantRepositoryImpl{db: db}
}

type trainingParticipantRow struct {
	ID         int64 `db:"id"`
	TrainingID int64 `db:"training_id"`
	MemberID   int64 `db:"member_id"`
}

// TrainingRepositoryImpl methods

func (r *TrainingRepositoryImpl) Create(ctx context.Context, training *entity.Training) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(
		ctx,
		`INSERT INTO trainings (club_id, location_id, date, duration, status, photo_file_id, message_id, created_at, updated_at)
		 VALUES (:club_id, :location_id, :date, :duration, :status, :photo_file_id, :message_id, :created_at, :updated_at)`,
		map[string]any{
			"club_id":       training.ClubID,
			"location_id":   training.LocationID,
			"date":          training.Date.Format(time.RFC3339),
			"duration":      training.Duration,
			"status":        training.Status,
			"photo_file_id": training.PhotoFileID,
			"message_id":    training.MessageID,
			"created_at":    now,
			"updated_at":    now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert training: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *TrainingRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Training, error) {
	var row trainingRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, club_id, location_id, date, duration, status, photo_file_id, message_id, created_at, updated_at FROM trainings WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("training not found: %w", err)
		}
		return nil, fmt.Errorf("get training by id: %w", err)
	}
	return trainingRowToEntity(&row), nil
}

func (r *TrainingRepositoryImpl) ListByClub(ctx context.Context, clubID int64) ([]entity.Training, error) {
	var rows []trainingRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, club_id, location_id, date, duration, status, photo_file_id, message_id, created_at, updated_at FROM trainings WHERE club_id = ? ORDER BY date DESC`,
		clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("list trainings by club: %w", err)
	}
	result := make([]entity.Training, len(rows))
	for i := range rows {
		result[i] = *trainingRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *TrainingRepositoryImpl) ListByStatus(
	ctx context.Context,
	status entity.TrainingStatus,
) ([]entity.Training, error) {
	var rows []trainingRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, club_id, location_id, date, duration, status, photo_file_id, message_id, created_at, updated_at FROM trainings WHERE status = ? ORDER BY date`,
		status,
	)
	if err != nil {
		return nil, fmt.Errorf("list trainings by status: %w", err)
	}
	result := make([]entity.Training, len(rows))
	for i := range rows {
		result[i] = *trainingRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *TrainingRepositoryImpl) Update(ctx context.Context, training *entity.Training) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE trainings SET club_id = :club_id, location_id = :location_id, date = :date,
		 duration = :duration, status = :status, photo_file_id = :photo_file_id,
		 message_id = :message_id, updated_at = :updated_at
		 WHERE id = :id`,
		map[string]any{
			"id":            training.ID,
			"club_id":       training.ClubID,
			"location_id":   training.LocationID,
			"date":          training.Date.Format(time.RFC3339),
			"duration":      training.Duration,
			"status":        training.Status,
			"photo_file_id": training.PhotoFileID,
			"message_id":    training.MessageID,
			"updated_at":    now,
		},
	)
	if err != nil {
		return fmt.Errorf("update training: %w", err)
	}
	return nil
}

func (r *TrainingRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM trainings WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete training: %w", err)
	}
	return nil
}

func trainingRowToEntity(row *trainingRow) *entity.Training {
	date, _ := time.Parse(time.RFC3339, row.Date)
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)
	return &entity.Training{
		ID:          row.ID,
		ClubID:      row.ClubID,
		LocationID:  row.LocationID,
		Date:        date,
		Duration:    row.Duration,
		Status:      entity.TrainingStatus(row.Status),
		PhotoFileID: row.PhotoFileID,
		MessageID:   row.MessageID,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

// TrainingTrainerRepositoryImpl methods

func (r *TrainingTrainerRepositoryImpl) Create(ctx context.Context, tt *entity.TrainingTrainer) (int64, error) {
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO training_trainers (training_id, member_id)
		 VALUES (:training_id, :member_id)`,
		map[string]any{
			"training_id": tt.TrainingID,
			"member_id":   tt.MemberID,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert training trainer: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *TrainingTrainerRepositoryImpl) ListByTraining(
	ctx context.Context,
	trainingID int64,
) ([]entity.TrainingTrainer, error) {
	var rows []trainingTrainerRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, training_id, member_id FROM training_trainers WHERE training_id = ?`, trainingID)
	if err != nil {
		return nil, fmt.Errorf("list training trainers: %w", err)
	}
	result := make([]entity.TrainingTrainer, len(rows))
	for i := range rows {
		result[i] = *trainingTrainerRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *TrainingTrainerRepositoryImpl) DeleteByTraining(ctx context.Context, trainingID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM training_trainers WHERE training_id = ?`, trainingID)
	if err != nil {
		return fmt.Errorf("delete training trainers: %w", err)
	}
	return nil
}

func trainingTrainerRowToEntity(row *trainingTrainerRow) *entity.TrainingTrainer {
	return &entity.TrainingTrainer{
		ID:         row.ID,
		TrainingID: row.TrainingID,
		MemberID:   row.MemberID,
	}
}

// TrainingParticipantRepositoryImpl methods

func (r *TrainingParticipantRepositoryImpl) Create(ctx context.Context, tp *entity.TrainingParticipant) (int64, error) {
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO training_participants (training_id, member_id)
		 VALUES (:training_id, :member_id)`,
		map[string]any{
			"training_id": tp.TrainingID,
			"member_id":   tp.MemberID,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert training participant: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *TrainingParticipantRepositoryImpl) GetByTrainingAndMember(
	ctx context.Context,
	trainingID, memberID int64,
) (*entity.TrainingParticipant, error) {
	var row trainingParticipantRow
	err := r.db.GetContext(ctx, &row,
		`SELECT id, training_id, member_id FROM training_participants WHERE training_id = ? AND member_id = ?`,
		trainingID, memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("training participant not found: %w", err)
		}
		return nil, fmt.Errorf("get training participant: %w", err)
	}
	return trainingParticipantRowToEntity(&row), nil
}

func (r *TrainingParticipantRepositoryImpl) ListByTraining(
	ctx context.Context,
	trainingID int64,
) ([]entity.TrainingParticipant, error) {
	var rows []trainingParticipantRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, training_id, member_id FROM training_participants WHERE training_id = ?`, trainingID)
	if err != nil {
		return nil, fmt.Errorf("list training participants: %w", err)
	}
	result := make([]entity.TrainingParticipant, len(rows))
	for i := range rows {
		result[i] = *trainingParticipantRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *TrainingParticipantRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM training_participants WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete training participant: %w", err)
	}
	return nil
}

func trainingParticipantRowToEntity(row *trainingParticipantRow) *entity.TrainingParticipant {
	return &entity.TrainingParticipant{
		ID:         row.ID,
		TrainingID: row.TrainingID,
		MemberID:   row.MemberID,
	}
}
