package handlers

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"runclub/internal/delivery/telegram/tgutil"
	"runclub/internal/usecase"
)

// RacePollHandler handles race registration callback queries.
type RacePollHandler struct {
	bot      *tgbotapi.BotAPI
	raceUC   usecase.RaceUseCase
	memberUC usecase.MemberUseCase
	logger   *zap.Logger
}

// NewRacePollHandler creates a new RacePollHandler.
func NewRacePollHandler(
	bot *tgbotapi.BotAPI,
	raceUC usecase.RaceUseCase,
	memberUC usecase.MemberUseCase,
	logger *zap.Logger,
) *RacePollHandler {
	return &RacePollHandler{
		bot:      bot,
		raceUC:   raceUC,
		memberUC: memberUC,
		logger:   logger,
	}
}

// HandleCallback processes race:reg:id:distance callbacks.
func (h *RacePollHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	data := tgutil.Decode(cb.Data)
	if len(data) < 4 {
		return nil
	}

	action := data[0]    // ActionRace
	subAction := data[1] // ActionRaceReg
	if action != ActionRace || subAction != ActionRaceReg {
		return nil
	}

	raceID, err := parseInt64(data[2])
	if err != nil {
		return fmt.Errorf("parse race ID: %w", err)
	}

	distance := data[3]

	// Auto-register the member if needed.
	member, err := h.memberUC.RegisterOrGet(ctx, cb.From.ID, cb.From.UserName)
	if err != nil {
		_ = answerCallback(h.bot, cb.ID, "Error: could not register you")
		return fmt.Errorf("register or get member: %w", err)
	}

	// Register the member for the race + distance.
	if err = h.raceUC.RegisterMember(ctx, raceID, member.ID, distance); err != nil {
		_ = answerCallback(h.bot, cb.ID, "Error registering for race")
		return fmt.Errorf("register member for race: %w", err)
	}

	_ = answerCallback(h.bot, cb.ID, fmt.Sprintf("Registered for %s!", distance))
	return nil
}
