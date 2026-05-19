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

type JointRunRepositoryImpl struct {
	db *sqlx.DB
}

func NewJointRunRepository(db *sqlx.DB) *JointRunRepositoryImpl {
	return &JointRunRepositoryImpl{db: db}
}

type jointRunRow struct {
	ID         int64  `db:"id"`
	ClubID     int64  `db:"club_id"`
	LocationID int64  `db:"location_id"`
	CreatorID  int64  `db:"creator_id"`
	Date       string `db:"date"`
	MessageID  int64  `db:"message_id"`
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

type JointRunParticipantRepositoryImpl struct {
	db *sqlx.DB
}

func NewJointRunParticipantRepository(db *sqlx.DB) *JointRunParticipantRepositoryImpl {
	return &JointRunParticipantRepositoryImpl{db: db}
}

type jointRunParticipantRow struct {
	ID         int64 `db:"id"`
	JointRunID int64 `db:"joint_run_id"`
	MemberID   int64 `db:"member_id"`
}

// JointRunRepositoryImpl methods

func (r *JointRunRepositoryImpl) Create(ctx context.Context, run *entity.JointRun) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO joint_runs (club_id, location_id, creator_id, date, message_id, created_at, updated_at)
		 VALUES (:club_id, :location_id, :creator_id, :date, :message_id, :created_at, :updated_at)`,
		map[string]any{
			"club_id":     run.ClubID,
			"location_id": run.LocationID,
			"creator_id":  run.CreatorID,
			"date":        run.Date.Format(time.RFC3339),
			"message_id":  run.MessageID,
			"created_at":  now,
			"updated_at":  now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert joint run: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *JointRunRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.JointRun, error) {
	var row jointRunRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, club_id, location_id, creator_id, date, message_id, created_at, updated_at FROM joint_runs WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("joint run not found: %w", err)
		}
		return nil, fmt.Errorf("get joint run by id: %w", err)
	}
	return jointRunRowToEntity(&row), nil
}

func (r *JointRunRepositoryImpl) ListByClub(ctx context.Context, clubID int64) ([]entity.JointRun, error) {
	var rows []jointRunRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, club_id, location_id, creator_id, date, message_id, created_at, updated_at FROM joint_runs WHERE club_id = ? ORDER BY date DESC`,
		clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("list joint runs by club: %w", err)
	}
	result := make([]entity.JointRun, len(rows))
	for i := range rows {
		result[i] = *jointRunRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *JointRunRepositoryImpl) Update(ctx context.Context, run *entity.JointRun) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE joint_runs SET club_id = :club_id, location_id = :location_id, creator_id = :creator_id,
		 date = :date, message_id = :message_id, updated_at = :updated_at
		 WHERE id = :id`,
		map[string]any{
			"id":          run.ID,
			"club_id":     run.ClubID,
			"location_id": run.LocationID,
			"creator_id":  run.CreatorID,
			"date":        run.Date.Format(time.RFC3339),
			"message_id":  run.MessageID,
			"updated_at":  now,
		},
	)
	if err != nil {
		return fmt.Errorf("update joint run: %w", err)
	}
	return nil
}

func (r *JointRunRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM joint_runs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete joint run: %w", err)
	}
	return nil
}

func jointRunRowToEntity(row *jointRunRow) *entity.JointRun {
	date, _ := time.Parse(time.RFC3339, row.Date)
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)
	return &entity.JointRun{
		ID:         row.ID,
		ClubID:     row.ClubID,
		LocationID: row.LocationID,
		CreatorID:  row.CreatorID,
		Date:       date,
		MessageID:  row.MessageID,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
}

// JointRunParticipantRepositoryImpl methods

func (r *JointRunParticipantRepositoryImpl) Create(ctx context.Context, p *entity.JointRunParticipant) (int64, error) {
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO joint_run_participants (joint_run_id, member_id)
		 VALUES (:joint_run_id, :member_id)`,
		map[string]any{
			"joint_run_id": p.JointRunID,
			"member_id":    p.MemberID,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert joint run participant: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *JointRunParticipantRepositoryImpl) GetByRunAndMember(
	ctx context.Context,
	runID, memberID int64,
) (*entity.JointRunParticipant, error) {
	var row jointRunParticipantRow
	err := r.db.GetContext(ctx, &row,
		`SELECT id, joint_run_id, member_id FROM joint_run_participants WHERE joint_run_id = ? AND member_id = ?`,
		runID, memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("joint run participant not found: %w", err)
		}
		return nil, fmt.Errorf("get joint run participant: %w", err)
	}
	return jointRunParticipantRowToEntity(&row), nil
}

func (r *JointRunParticipantRepositoryImpl) ListByRun(
	ctx context.Context,
	runID int64,
) ([]entity.JointRunParticipant, error) {
	var rows []jointRunParticipantRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, joint_run_id, member_id FROM joint_run_participants WHERE joint_run_id = ?`, runID)
	if err != nil {
		return nil, fmt.Errorf("list joint run participants: %w", err)
	}
	result := make([]entity.JointRunParticipant, len(rows))
	for i := range rows {
		result[i] = *jointRunParticipantRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *JointRunParticipantRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM joint_run_participants WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete joint run participant: %w", err)
	}
	return nil
}

func jointRunParticipantRowToEntity(row *jointRunParticipantRow) *entity.JointRunParticipant {
	return &entity.JointRunParticipant{
		ID:         row.ID,
		JointRunID: row.JointRunID,
		MemberID:   row.MemberID,
	}
}
