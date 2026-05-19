package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"runclub/internal/delivery/telegram/tgutil"
	"runclub/internal/domain/entity"
	"runclub/internal/pkg/tgrescape"
	"runclub/internal/usecase"
)

// ParticipationHandler handles "Иду"/"Не иду" button toggles for trainings and joint runs.
type ParticipationHandler struct {
	bot        *tgbotapi.BotAPI
	trainingUC usecase.TrainingUseCase
	jointRunUC usecase.JointRunUseCase
	memberUC   usecase.MemberUseCase
	clubUC     usecase.ClubUseCase
	logger     *zap.Logger
}

// NewParticipationHandler creates a new ParticipationHandler.
func NewParticipationHandler(
	bot *tgbotapi.BotAPI,
	trainingUC usecase.TrainingUseCase,
	jointRunUC usecase.JointRunUseCase,
	memberUC usecase.MemberUseCase,
	clubUC usecase.ClubUseCase,
	logger *zap.Logger,
) *ParticipationHandler {
	return &ParticipationHandler{
		bot:        bot,
		trainingUC: trainingUC,
		jointRunUC: jointRunUC,
		memberUC:   memberUC,
		clubUC:     clubUC,
		logger:     logger,
	}
}

// HandleCallback processes join/leave callbacks for trainings and joint runs.
func (h *ParticipationHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	data := tgutil.Decode(cb.Data)
	if len(data) < 3 {
		return nil
	}

	action := data[0]     // ActionJoin or ActionLeave
	entityType := data[1] // EntityTraining or EntityJointRun
	idStr := data[2]

	entityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("parse entity ID: %w", err)
	}

	// Auto-register member if they don't exist.
	member, err := h.memberUC.RegisterOrGet(ctx, cb.From.ID, cb.From.UserName)
	if err != nil {
		_ = answerCallback(h.bot, cb.ID, "Error: could not register you")
		return fmt.Errorf("register or get member: %w", err)
	}

	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	switch entityType {
	case EntityTraining:
		return h.handleTrainingParticipation(ctx, cb, action, entityID, member, chatID, messageID)
	case EntityJointRun:
		return h.handleJointRunParticipation(ctx, cb, action, entityID, member, chatID, messageID)
	default:
		return nil
	}
}

// handleTrainingParticipation processes join/leave for a training.
func (h *ParticipationHandler) handleTrainingParticipation(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	action string,
	trainingID int64,
	member *entity.Member,
	chatID int64,
	messageID int,
) error {
	isJoin := action == ActionJoin

	if err := h.toggleParticipant(ctx, isJoin, trainingID, member.ID, cb,
		h.trainingUC.AddParticipant, h.trainingUC.RemoveParticipant,
		"You left the training",
	); err != nil {
		return err
	}

	return h.updateTrainingMessage(ctx, trainingID, chatID, messageID, member.ID, isJoin)
}

// handleJointRunParticipation processes join/leave for a joint run.
func (h *ParticipationHandler) handleJointRunParticipation(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	action string,
	runID int64,
	member *entity.Member,
	chatID int64,
	messageID int,
) error {
	isJoin := action == ActionJoin

	if err := h.toggleParticipant(ctx, isJoin, runID, member.ID, cb,
		h.jointRunUC.AddParticipant, h.jointRunUC.RemoveParticipant,
		"You left the run",
	); err != nil {
		return err
	}

	return h.updateJointRunMessage(ctx, runID, chatID, messageID, member.ID, isJoin)
}

type addParticipantFunc func(ctx context.Context, entityID, memberID int64) error
type removeParticipantFunc func(ctx context.Context, entityID, memberID int64) error

func (h *ParticipationHandler) toggleParticipant(
	ctx context.Context,
	isJoin bool,
	entityID int64,
	memberID int64,
	cb *tgbotapi.CallbackQuery,
	addFn addParticipantFunc,
	removeFn removeParticipantFunc,
	leaveMsg string,
) error {
	if isJoin {
		if err := addFn(ctx, entityID, memberID); err != nil {
			_ = answerCallback(h.bot, cb.ID, "Error joining")
			return fmt.Errorf("add participant: %w", err)
		}
		_ = answerCallback(h.bot, cb.ID, "You joined!")
	} else {
		if err := removeFn(ctx, entityID, memberID); err != nil {
			_ = answerCallback(h.bot, cb.ID, "Error leaving")
			return fmt.Errorf("remove participant: %w", err)
		}
		_ = answerCallback(h.bot, cb.ID, leaveMsg)
	}
	return nil
}

// updateTrainingMessage edits the training message to show the updated participant list.
func (h *ParticipationHandler) updateTrainingMessage(
	ctx context.Context,
	trainingID int64,
	chatID int64,
	messageID int,
	_ int64,
	isJoined bool,
) error {
	participants, err := h.trainingUC.ListParticipants(ctx, trainingID)
	if err != nil {
		return err
	}

	text := buildParticipantList("Training", participants)
	keyboard := tgutil.ParticipationKeyboard(trainingID, isJoined)

	return editMessageText(h.bot, chatID, messageID, text, &keyboard)
}

// updateJointRunMessage edits the joint run message to show the updated participant list.
func (h *ParticipationHandler) updateJointRunMessage(
	ctx context.Context,
	runID int64,
	chatID int64,
	messageID int,
	_ int64,
	isJoined bool,
) error {
	participants, err := h.jointRunUC.ListParticipants(ctx, runID)
	if err != nil {
		return err
	}

	text := buildParticipantList("Joint Run", participants)
	keyboard := tgutil.ParticipationKeyboardJointRun(runID, isJoined)

	return editMessageText(h.bot, chatID, messageID, text, &keyboard)
}

// buildParticipantList builds a MarkdownV2 message text with the participant list.
func buildParticipantList(title string, participants []entity.Member) string {
	text := tgrescape.EscapeMarkdownV2(title) + "\n\n"
	if len(participants) == 0 {
		text += tgrescape.EscapeMarkdownV2("No participants yet.")
	} else {
		text += tgrescape.EscapeMarkdownV2("Participants:") + "\n"
		var textSb189 strings.Builder
		for i, p := range participants {
			fmt.Fprintf(&textSb189, "%d\\. %s\n", i+1, tgrescape.EscapeMarkdownV2(tgutil.DisplayName(p)))
		}
		text += textSb189.String()
	}
	return text
}
