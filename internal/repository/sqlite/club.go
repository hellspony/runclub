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

type ClubRepositoryImpl struct {
	db *sqlx.DB
}

func NewClubRepository(db *sqlx.DB) *ClubRepositoryImpl {
	return &ClubRepositoryImpl{db: db}
}

type clubRow struct {
	ID                int64  `db:"id"`
	Name              string `db:"name"`
	TelegramChatID    int64  `db:"telegram_chat_id"`
	WelcomeEnabled    bool   `db:"welcome_enabled"`
	BirthdayEnabled   bool   `db:"birthday_enabled"`
	RaceNotifyEnabled bool   `db:"race_notify_enabled"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
}

func (r *ClubRepositoryImpl) Create(ctx context.Context, club *entity.Club) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(
		ctx,
		`INSERT INTO clubs (name, telegram_chat_id, welcome_enabled, birthday_enabled, race_notify_enabled, created_at, updated_at)
		 VALUES (:name, :telegram_chat_id, :welcome_enabled, :birthday_enabled, :race_notify_enabled, :created_at, :updated_at)`,
		map[string]any{
			"name":                club.Name,
			"telegram_chat_id":    club.TelegramChatID,
			"welcome_enabled":     club.WelcomeEnabled,
			"birthday_enabled":    club.BirthdayEnabled,
			"race_notify_enabled": club.RaceNotifyEnabled,
			"created_at":          now,
			"updated_at":          now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert club: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *ClubRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Club, error) {
	var row clubRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, name, telegram_chat_id, welcome_enabled, birthday_enabled, race_notify_enabled, created_at, updated_at FROM clubs WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("club not found: %w", err)
		}
		return nil, fmt.Errorf("get club by id: %w", err)
	}
	return clubRowToEntity(&row), nil
}

func (r *ClubRepositoryImpl) GetByTelegramChatID(ctx context.Context, chatID int64) (*entity.Club, error) {
	var row clubRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, name, telegram_chat_id, welcome_enabled, birthday_enabled, race_notify_enabled, created_at, updated_at FROM clubs WHERE telegram_chat_id = ?`,
		chatID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("club not found: %w", err)
		}
		return nil, fmt.Errorf("get club by telegram chat id: %w", err)
	}
	return clubRowToEntity(&row), nil
}

func (r *ClubRepositoryImpl) List(ctx context.Context) ([]entity.Club, error) {
	var rows []clubRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, name, telegram_chat_id, welcome_enabled, birthday_enabled, race_notify_enabled, created_at, updated_at FROM clubs ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	result := make([]entity.Club, len(rows))
	for i := range rows {
		result[i] = *clubRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *ClubRepositoryImpl) Update(ctx context.Context, club *entity.Club) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE clubs SET name = :name, telegram_chat_id = :telegram_chat_id,
		 welcome_enabled = :welcome_enabled, birthday_enabled = :birthday_enabled,
		 race_notify_enabled = :race_notify_enabled, updated_at = :updated_at
		 WHERE id = :id`,
		map[string]any{
			"id":                  club.ID,
			"name":                club.Name,
			"telegram_chat_id":    club.TelegramChatID,
			"welcome_enabled":     club.WelcomeEnabled,
			"birthday_enabled":    club.BirthdayEnabled,
			"race_notify_enabled": club.RaceNotifyEnabled,
			"updated_at":          now,
		},
	)
	if err != nil {
		return fmt.Errorf("update club: %w", err)
	}
	return nil
}

func (r *ClubRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM clubs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete club: %w", err)
	}
	return nil
}

func clubRowToEntity(row *clubRow) *entity.Club {
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)
	return &entity.Club{
		ID:                row.ID,
		Name:              row.Name,
		TelegramChatID:    row.TelegramChatID,
		WelcomeEnabled:    row.WelcomeEnabled,
		BirthdayEnabled:   row.BirthdayEnabled,
		RaceNotifyEnabled: row.RaceNotifyEnabled,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}
}
