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

type BotStateRepositoryImpl struct {
	db *sqlx.DB
}

func NewBotStateRepository(db *sqlx.DB) *BotStateRepositoryImpl {
	return &BotStateRepositoryImpl{db: db}
}

type botStateRow struct {
	ID         int64  `db:"id"`
	TelegramID int64  `db:"telegram_id"`
	ChatID     int64  `db:"chat_id"`
	Flow       string `db:"flow"`
	Step       int    `db:"step"`
	Payload    string `db:"payload"`
	UpdatedAt  string `db:"updated_at"`
}

func (r *BotStateRepositoryImpl) Create(ctx context.Context, state *entity.BotState) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO bot_states (telegram_id, chat_id, flow, step, payload, updated_at)
		 VALUES (:telegram_id, :chat_id, :flow, :step, :payload, :updated_at)`,
		map[string]any{
			"telegram_id": state.TelegramID,
			"chat_id":     state.ChatID,
			"flow":        state.Flow,
			"step":        state.Step,
			"payload":     state.Payload,
			"updated_at":  now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert bot state: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *BotStateRepositoryImpl) GetByTelegramAndFlow(
	ctx context.Context,
	telegramID, chatID int64,
	flow entity.BotFlowType,
) (*entity.BotState, error) {
	var row botStateRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, telegram_id, chat_id, flow, step, payload, updated_at FROM bot_states WHERE telegram_id = ? AND chat_id = ? AND flow = ?`,
		telegramID,
		chatID,
		flow,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("bot state not found: %w", err)
		}
		return nil, fmt.Errorf("get bot state: %w", err)
	}
	return botStateRowToEntity(&row), nil
}

func (r *BotStateRepositoryImpl) Update(ctx context.Context, state *entity.BotState) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE bot_states SET telegram_id = :telegram_id, chat_id = :chat_id,
		 flow = :flow, step = :step, payload = :payload, updated_at = :updated_at
		 WHERE id = :id`,
		map[string]any{
			"id":          state.ID,
			"telegram_id": state.TelegramID,
			"chat_id":     state.ChatID,
			"flow":        state.Flow,
			"step":        state.Step,
			"payload":     state.Payload,
			"updated_at":  now,
		},
	)
	if err != nil {
		return fmt.Errorf("update bot state: %w", err)
	}
	return nil
}

func (r *BotStateRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM bot_states WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete bot state: %w", err)
	}
	return nil
}

func (r *BotStateRepositoryImpl) DeleteOlderThan(ctx context.Context, before time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM bot_states WHERE updated_at < ?`, before.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("delete older bot states: %w", err)
	}
	return nil
}

func botStateRowToEntity(row *botStateRow) *entity.BotState {
	updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)
	return &entity.BotState{
		ID:         row.ID,
		TelegramID: row.TelegramID,
		ChatID:     row.ChatID,
		Flow:       entity.BotFlowType(row.Flow),
		Step:       row.Step,
		Payload:    row.Payload,
		UpdatedAt:  updatedAt,
	}
}
