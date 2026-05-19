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

type MemberRepositoryImpl struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) *MemberRepositoryImpl {
	return &MemberRepositoryImpl{db: db}
}

type memberRow struct {
	ID               int64          `db:"id"`
	FIO              string         `db:"fio"`
	TelegramUsername string         `db:"telegram_username"`
	TelegramID       int64          `db:"telegram_id"`
	BirthDate        sql.NullString `db:"birth_date"`
	LeftAt           sql.NullString `db:"left_at"`
	CreatedAt        string         `db:"created_at"`
	UpdatedAt        string         `db:"updated_at"`
}

type ClubMemberRepositoryImpl struct {
	db *sqlx.DB
}

func NewClubMemberRepository(db *sqlx.DB) *ClubMemberRepositoryImpl {
	return &ClubMemberRepositoryImpl{db: db}
}

type clubMemberRow struct {
	ID       int64  `db:"id"`
	ClubID   int64  `db:"club_id"`
	MemberID int64  `db:"member_id"`
	Role     string `db:"role"`
	JoinedAt string `db:"joined_at"`
}

func (r *MemberRepositoryImpl) Create(ctx context.Context, member *entity.Member) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	var birthDate any
	if member.BirthDate != nil {
		birthDate = member.BirthDate.Format(time.RFC3339)
	}
	var leftAt any
	if member.LeftAt != nil {
		leftAt = member.LeftAt.Format(time.RFC3339)
	}
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO members (fio, telegram_username, telegram_id, birth_date, left_at, created_at, updated_at)
		 VALUES (:fio, :telegram_username, :telegram_id, :birth_date, :left_at, :created_at, :updated_at)`,
		map[string]any{
			"fio":               member.FIO,
			"telegram_username": member.TelegramUsername,
			"telegram_id":       member.TelegramID,
			"birth_date":        birthDate,
			"left_at":           leftAt,
			"created_at":        now,
			"updated_at":        now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert member: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *MemberRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Member, error) {
	var row memberRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, fio, telegram_username, telegram_id, birth_date, created_at, updated_at, left_at FROM members WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("member not found: %w", err)
		}
		return nil, fmt.Errorf("get member by id: %w", err)
	}
	return memberRowToEntity(&row), nil
}

func (r *MemberRepositoryImpl) GetByTelegramID(ctx context.Context, telegramID int64) (*entity.Member, error) {
	var row memberRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, fio, telegram_username, telegram_id, birth_date, created_at, updated_at, left_at FROM members WHERE telegram_id = ?`,
		telegramID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("member not found: %w", err)
		}
		return nil, fmt.Errorf("get member by telegram id: %w", err)
	}
	return memberRowToEntity(&row), nil
}

func (r *MemberRepositoryImpl) ListByClub(ctx context.Context, clubID int64) ([]entity.Member, error) {
	var rows []memberRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT m.id, m.fio, m.telegram_username, m.telegram_id, m.birth_date, m.created_at, m.updated_at, m.left_at FROM members m
		 JOIN club_members cm ON cm.member_id = m.id
		 WHERE cm.club_id = ?
		 ORDER BY m.fio`,
		clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("list members by club: %w", err)
	}
	result := make([]entity.Member, len(rows))
	for i := range rows {
		result[i] = *memberRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *MemberRepositoryImpl) ListTrainersByClub(ctx context.Context, clubID int64) ([]entity.Member, error) {
	var rows []memberRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT m.id, m.fio, m.telegram_username, m.telegram_id, m.birth_date, m.created_at, m.updated_at, m.left_at FROM members m
		 JOIN club_members cm ON cm.member_id = m.id
		 WHERE cm.club_id = ? AND cm.role = 'trainer'
		 ORDER BY m.fio`,
		clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("list trainers by club: %w", err)
	}
	result := make([]entity.Member, len(rows))
	for i := range rows {
		result[i] = *memberRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *MemberRepositoryImpl) ListBirthdayOn(ctx context.Context, month, day int) ([]entity.Member, error) {
	var rows []memberRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, fio, telegram_username, telegram_id, birth_date, created_at, updated_at, left_at FROM members
		 WHERE CAST(strftime('%m', birth_date) AS INTEGER) = ? AND CAST(strftime('%d', birth_date) AS INTEGER) = ?`,
		month, day)
	if err != nil {
		return nil, fmt.Errorf("list birthday on: %w", err)
	}
	result := make([]entity.Member, len(rows))
	for i := range rows {
		result[i] = *memberRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *MemberRepositoryImpl) Update(ctx context.Context, member *entity.Member) error {
	now := time.Now().Format(time.RFC3339)
	var birthDate any
	if member.BirthDate != nil {
		birthDate = member.BirthDate.Format(time.RFC3339)
	}
	var leftAt any
	if member.LeftAt != nil {
		leftAt = member.LeftAt.Format(time.RFC3339)
	}
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE members SET fio = :fio, telegram_username = :telegram_username,
		 telegram_id = :telegram_id, birth_date = :birth_date, left_at = :left_at, updated_at = :updated_at
		 WHERE id = :id`,
		map[string]any{
			"id":                member.ID,
			"fio":               member.FIO,
			"telegram_username": member.TelegramUsername,
			"telegram_id":       member.TelegramID,
			"birth_date":        birthDate,
			"left_at":           leftAt,
			"updated_at":        now,
		},
	)
	if err != nil {
		return fmt.Errorf("update member: %w", err)
	}
	return nil
}

func (r *MemberRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM members WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete member: %w", err)
	}
	return nil
}

func (r *MemberRepositoryImpl) ListOrphansOlderThan(ctx context.Context, days int) ([]entity.Member, error) {
	var rows []memberRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT m.id, m.fio, m.telegram_username, m.telegram_id, m.birth_date, m.created_at, m.updated_at, m.left_at FROM members m
		 WHERE m.id NOT IN (SELECT member_id FROM club_members)
		   AND m.left_at IS NOT NULL
		   AND datetime(m.left_at) < datetime('now', '-' || ? || ' days')`,
		days,
	)
	if err != nil {
		return nil, fmt.Errorf("list orphan members: %w", err)
	}
	result := make([]entity.Member, len(rows))
	for i := range rows {
		result[i] = *memberRowToEntity(&rows[i])
	}
	return result, nil
}

func memberRowToEntity(row *memberRow) *entity.Member {
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)
	var birthDate *time.Time
	if row.BirthDate.Valid && row.BirthDate.String != "" {
		t, err := time.Parse(time.RFC3339, row.BirthDate.String)
		if err == nil {
			birthDate = &t
		}
	}
	var leftAt *time.Time
	if row.LeftAt.Valid && row.LeftAt.String != "" {
		t, err := time.Parse(time.RFC3339, row.LeftAt.String)
		if err == nil {
			leftAt = &t
		}
	}
	return &entity.Member{
		ID:               row.ID,
		FIO:              row.FIO,
		TelegramUsername: row.TelegramUsername,
		TelegramID:       row.TelegramID,
		BirthDate:        birthDate,
		LeftAt:           leftAt,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
}

// ClubMemberRepositoryImpl methods

func (r *ClubMemberRepositoryImpl) Create(ctx context.Context, cm *entity.ClubMember) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO club_members (club_id, member_id, role, joined_at)
		 VALUES (:club_id, :member_id, :role, :joined_at)`,
		map[string]any{
			"club_id":   cm.ClubID,
			"member_id": cm.MemberID,
			"role":      cm.Role,
			"joined_at": now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert club member: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *ClubMemberRepositoryImpl) GetByClubAndMember(
	ctx context.Context,
	clubID, memberID int64,
) (*entity.ClubMember, error) {
	var row clubMemberRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, club_id, member_id, role, joined_at FROM club_members WHERE club_id = ? AND member_id = ?`,
		clubID,
		memberID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("club member not found: %w", err)
		}
		return nil, fmt.Errorf("get club member: %w", err)
	}
	return clubMemberRowToEntity(&row), nil
}

func (r *ClubMemberRepositoryImpl) ListClubsByMember(ctx context.Context, memberID int64) ([]entity.ClubMember, error) {
	var rows []clubMemberRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, club_id, member_id, role, joined_at FROM club_members WHERE member_id = ?`, memberID)
	if err != nil {
		return nil, fmt.Errorf("list clubs by member: %w", err)
	}
	result := make([]entity.ClubMember, len(rows))
	for i := range rows {
		result[i] = *clubMemberRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *ClubMemberRepositoryImpl) ListTrainerClubs(ctx context.Context, memberID int64) ([]entity.ClubMember, error) {
	var rows []clubMemberRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, club_id, member_id, role, joined_at FROM club_members WHERE member_id = ? AND role = 'trainer'`,
		memberID,
	)
	if err != nil {
		return nil, fmt.Errorf("list trainer clubs: %w", err)
	}
	result := make([]entity.ClubMember, len(rows))
	for i := range rows {
		result[i] = *clubMemberRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *ClubMemberRepositoryImpl) UpdateRole(
	ctx context.Context,
	clubID, memberID int64,
	role entity.MemberRole,
) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE club_members SET role = ? WHERE club_id = ? AND member_id = ?`,
		role, clubID, memberID)
	if err != nil {
		return fmt.Errorf("update club member role: %w", err)
	}
	return nil
}

func (r *ClubMemberRepositoryImpl) Delete(ctx context.Context, clubID, memberID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM club_members WHERE club_id = ? AND member_id = ?`, clubID, memberID)
	if err != nil {
		return fmt.Errorf("delete club member: %w", err)
	}
	return nil
}

func clubMemberRowToEntity(row *clubMemberRow) *entity.ClubMember {
	joinedAt, _ := time.Parse(time.RFC3339, row.JoinedAt)
	return &entity.ClubMember{
		ID:       row.ID,
		ClubID:   row.ClubID,
		MemberID: row.MemberID,
		Role:     entity.MemberRole(row.Role),
		JoinedAt: joinedAt,
	}
}
