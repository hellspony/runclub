package repository

import (
	"context"
	"time"

	"runclub/internal/domain/entity"
)

type BotStateRepository interface {
	Create(ctx context.Context, state *entity.BotState) (int64, error)
	GetByTelegramAndFlow(
		ctx context.Context,
		telegramID, chatID int64,
		flow entity.BotFlowType,
	) (*entity.BotState, error)
	Update(ctx context.Context, state *entity.BotState) error
	Delete(ctx context.Context, id int64) error
	DeleteOlderThan(ctx context.Context, before time.Time) error
}
