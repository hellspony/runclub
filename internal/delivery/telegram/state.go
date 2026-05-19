package telegram

import (
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

// StepFunc is a handler for a single step of a bot flow.
type StepFunc func(ctx context.Context, msg *tgbotapi.Message, state *entity.BotState) error

// StateMachine manages multi-step bot flows (FSM).
type StateMachine struct {
	repo   repository.BotStateRepository
	steps  map[entity.BotFlowType]map[int]StepFunc
	logger *zap.Logger
}

// NewStateMachine creates a new StateMachine.
func NewStateMachine(repo repository.BotStateRepository, logger *zap.Logger) *StateMachine {
	return &StateMachine{
		repo:   repo,
		steps:  make(map[entity.BotFlowType]map[int]StepFunc),
		logger: logger,
	}
}

// RegisterStep registers a step handler for a given flow type and step number.
func (sm *StateMachine) RegisterStep(flow entity.BotFlowType, step int, fn StepFunc) {
	if sm.steps[flow] == nil {
		sm.steps[flow] = make(map[int]StepFunc)
	}
	sm.steps[flow][step] = fn
}

// CreateState creates a new bot state in the repository.
func (sm *StateMachine) CreateState(ctx context.Context, state *entity.BotState) error {
	_, err := sm.repo.Create(ctx, state)
	if err != nil {
		return fmt.Errorf("create bot state: %w", err)
	}
	return nil
}

// GetState retrieves an active bot state by telegram ID, chat ID, and flow type.
func (sm *StateMachine) GetState(
	ctx context.Context,
	telegramID, chatID int64,
	flow entity.BotFlowType,
) (*entity.BotState, error) {
	state, err := sm.repo.GetByTelegramAndFlow(ctx, telegramID, chatID, flow)
	if err != nil {
		return nil, fmt.Errorf("get bot state: %w", err)
	}
	return state, nil
}

// SetStep updates the step number for the given bot state.
func (sm *StateMachine) SetStep(ctx context.Context, state *entity.BotState, step int) error {
	state.Step = step
	if err := sm.repo.Update(ctx, state); err != nil {
		return fmt.Errorf("update bot state step: %w", err)
	}
	return nil
}

// UpdatePayload serializes and updates the payload for the given bot state.
func (sm *StateMachine) UpdatePayload(ctx context.Context, state *entity.BotState, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	state.Payload = string(data)
	if err = sm.repo.Update(ctx, state); err != nil {
		return fmt.Errorf("update bot state payload: %w", err)
	}
	return nil
}

// ClearState deletes the bot state from the repository.
func (sm *StateMachine) ClearState(ctx context.Context, state *entity.BotState) error {
	if err := sm.repo.Delete(ctx, state.ID); err != nil {
		return fmt.Errorf("delete bot state: %w", err)
	}
	return nil
}

// Process looks up the active state for the message sender and dispatches to the step handler.
func (sm *StateMachine) Process(ctx context.Context, msg *tgbotapi.Message) error {
	if msg == nil || msg.From == nil {
		return nil
	}

	telegramID := msg.From.ID
	chatID := msg.Chat.ID

	// Try all registered flow types to find an active state.
	for flow := range sm.steps {
		state, err := sm.repo.GetByTelegramAndFlow(ctx, telegramID, chatID, flow)
		if err != nil {
			// No active state for this flow, try next.
			continue
		}

		stepHandlers, ok := sm.steps[flow]
		if !ok {
			continue
		}

		handler, ok := stepHandlers[state.Step]
		if !ok {
			sm.logger.Warn("no handler for step",
				zap.String("flow", string(flow)),
				zap.Int("step", state.Step),
			)
			continue
		}

		return handler(ctx, msg, state)
	}

	return nil
}

// HasActiveState checks if there is an active state for a user in a chat for any flow.
func (sm *StateMachine) HasActiveState(ctx context.Context, telegramID, chatID int64) bool {
	for flow := range sm.steps {
		_, err := sm.repo.GetByTelegramAndFlow(ctx, telegramID, chatID, flow)
		if err == nil {
			return true
		}
	}
	return false
}
