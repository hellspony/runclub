package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"runclub/internal/delivery/telegram/tgutil"
	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
	"runclub/internal/pkg/templater"
	"runclub/internal/pkg/tgrescape"
	"runclub/internal/usecase"
)

// ConfirmationHandler handles post-training confirmation flows.
type ConfirmationHandler struct {
	bot          *tgbotapi.BotAPI
	trainingUC   usecase.TrainingUseCase
	memberUC     usecase.MemberUseCase
	clubUC       usecase.ClubUseCase
	templateUC   usecase.TemplateUseCase
	locationUC   usecase.LocationUseCase
	botStateRepo repository.BotStateRepository
	logger       *zap.Logger
}

// NewConfirmationHandler creates a new ConfirmationHandler.
func NewConfirmationHandler(
	bot *tgbotapi.BotAPI,
	trainingUC usecase.TrainingUseCase,
	memberUC usecase.MemberUseCase,
	clubUC usecase.ClubUseCase,
	templateUC usecase.TemplateUseCase,
	locationUC usecase.LocationUseCase,
	botStateRepo repository.BotStateRepository,
	logger *zap.Logger,
) *ConfirmationHandler {
	return &ConfirmationHandler{
		bot:          bot,
		trainingUC:   trainingUC,
		memberUC:     memberUC,
		clubUC:       clubUC,
		templateUC:   templateUC,
		locationUC:   locationUC,
		botStateRepo: botStateRepo,
		logger:       logger,
	}
}

// SendConfirmationPrompt sends a DM to the trainer with the participant list and confirmation keyboard.
func (h *ConfirmationHandler) SendConfirmationPrompt(ctx context.Context, trainingID int64) error {
	_, err := h.trainingUC.GetTraining(ctx, trainingID)
	if err != nil {
		return fmt.Errorf("get training: %w", err)
	}

	// Start the confirmation process - update status.
	if err = h.trainingUC.StartConfirmation(ctx, trainingID); err != nil {
		return fmt.Errorf("start confirmation: %w", err)
	}

	// Get trainers for this training.
	trainers, err := h.trainingUC.ListTrainers(ctx, trainingID)
	if err != nil {
		return fmt.Errorf("list trainers: %w", err)
	}

	// Get participants.
	participants, err := h.trainingUC.ListParticipants(ctx, trainingID)
	if err != nil {
		return fmt.Errorf("list participants: %w", err)
	}

	// Build participant list text.
	var participantList string
	var participantListSb80 strings.Builder
	for i, p := range participants {
		fmt.Fprintf(&participantListSb80, "%d. %s\n", i+1, tgutil.DisplayName(p))
	}
	participantList += participantListSb80.String()
	if participantList == "" {
		participantList = "No participants yet."
	}

	text := fmt.Sprintf(
		"Training #%d needs confirmation.\n\nParticipants:\n%s\nPlease confirm or edit the participant list.",
		trainingID,
		participantList,
	)

	keyboard := tgutil.ConfirmationKeyboard(trainingID)

	// Send DM to each trainer.
	for _, trainer := range trainers {
		// Open a private chat with the trainer.
		dmChat := tgbotapi.NewMessage(trainer.TelegramID, text)
		// Telegram ID is not a chat ID for DMs. We need to use the TelegramID directly.
		// Unfortunately, bots can only initiate DMs if the user has previously messaged the bot.
		// We'll try to send to the trainer's Telegram ID as a chat ID (works for DMs).
		dmChat.ChatID = trainer.TelegramID
		dmChat.ReplyMarkup = keyboard
		_, sendErr := h.bot.Send(dmChat)
		if sendErr != nil {
			h.logger.Error("failed to send confirmation DM to trainer",
				zap.Int64("trainer_id", trainer.ID),
				zap.Error(sendErr),
			)
			continue
		}
	}

	return nil
}

// HandleCallback processes confirmation-related callback queries.
func (h *ConfirmationHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	data := tgutil.Decode(cb.Data)
	if len(data) < 3 {
		return nil
	}

	action := data[0]     // ActionConfirm, ActionPhoto, ActionAddSelect, ActionRemSelect
	entityType := data[1] // EntityTraining
	if entityType != EntityTraining {
		return nil
	}

	trainingID, err := parseInt64(data[2])
	if err != nil {
		return err
	}

	// Only process if this is a confirmation flow (not training creation).
	// Training creation uses "addsel"/"remsel" with "trainer" entity type,
	// while confirmation uses "training" entity type.
	state, _ := h.botStateRepo.GetByTelegramAndFlow(ctx, cb.From.ID, cb.Message.Chat.ID, entity.FlowTrainingConfirm)

	switch action {
	case ActionConfirm:
		return h.handleConfirm(ctx, cb, trainingID, state)
	case ActionPhoto:
		return h.handlePhoto(ctx, cb, trainingID, state)
	case ActionAddSelect:
		return h.handleAddParticipant(ctx, cb, trainingID, state)
	case ActionRemSelect:
		return h.handleRemoveParticipant(ctx, cb, trainingID, state)
	default:
		return nil
	}
}

// handleConfirm confirms the training and posts the final message.
func (h *ConfirmationHandler) handleConfirm(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	trainingID int64,
	state *entity.BotState,
) error {
	_ = answerCallback(h.bot, cb.ID, "Confirmed!")

	var addedIDs, removedIDs []int64
	var photoFileID string

	if state != nil {
		var payload entity.TrainingConfirmPayload
		if err := json.Unmarshal([]byte(state.Payload), &payload); err == nil {
			addedIDs = payload.AddedIDs
			removedIDs = payload.RemovedIDs
			photoFileID = payload.PhotoFileID
		}
		_ = h.botStateRepo.Delete(ctx, state.ID)
	}

	// Confirm the training in the use case.
	if err := h.trainingUC.ConfirmTraining(ctx, trainingID, addedIDs, removedIDs, photoFileID); err != nil {
		return fmt.Errorf("confirm training: %w", err)
	}

	// Get training details.
	training, err := h.trainingUC.GetTraining(ctx, trainingID)
	if err != nil {
		return err
	}

	// Get participants (including additions/removals).
	participants, err := h.trainingUC.ListParticipants(ctx, trainingID)
	if err != nil {
		return err
	}

	// Get trainers.
	trainers, err := h.trainingUC.ListTrainers(ctx, trainingID)
	if err != nil {
		return err
	}

	// Get location.
	location, err := h.locationUC.GetByID(ctx, training.LocationID)
	if err != nil {
		return err
	}

	// Get club.
	club, err := h.clubUC.GetByID(ctx, training.ClubID)
	if err != nil {
		return err
	}

	// Build participant and trainer name lists.
	var participantNames []string
	for _, p := range participants {
		participantNames = append(participantNames, tgutil.DisplayName(p))
	}
	var trainerNames []string
	for _, t := range trainers {
		trainerNames = append(trainerNames, tgutil.DisplayName(t))
	}

	// Render template.
	tmpl, err := h.templateUC.GetByClubAndType(ctx, training.ClubID, entity.TemplateTrainingDone)
	var text string
	if err != nil || tmpl == nil {
		text = fmt.Sprintf("Training completed: %s, %s, %d min\nTrainers: %v\nParticipants: %v",
			tgrescape.EscapeMarkdownV2(location.Name),
			tgrescape.EscapeMarkdownV2(training.Date.Format("02.01.2006 15:04")),
			training.Duration,
			trainerNames,
			participantNames,
		)
	} else {
		text, err = templater.Render(tmpl.Content, map[string]any{
			TplLocationName: location.Name,
			TplDate:         training.Date.Format("02.01.2006 15:04"),
			"Duration":      training.Duration,
			"Trainers":      trainerNames,
			"Participants":  participantNames,
			"PhotoFileID":   photoFileID,
		})
		if err != nil {
			text = "Training completed!"
		}
		text = tgrescape.EscapeMarkdownV2(text)
	}

	// Post to club chat.
	chatMsg := tgbotapi.NewMessage(club.TelegramChatID, text)
	chatMsg.ParseMode = tgbotapi.ModeMarkdownV2

	// Attach photo if available.
	if photoFileID != "" {
		photoMsg := tgbotapi.NewPhoto(club.TelegramChatID, tgbotapi.FileID(photoFileID))
		photoMsg.Caption = text
		photoMsg.ParseMode = tgbotapi.ModeMarkdownV2
		_, err = h.bot.Send(photoMsg)
	} else {
		_, err = h.bot.Send(chatMsg)
	}

	if err != nil {
		return fmt.Errorf("post confirmed training: %w", err)
	}

	_ = sendText(h.bot, cb.Message.Chat.ID, "Training confirmed and posted to the club chat\\!")
	return nil
}

// handlePhoto sets the FSM to expect a photo upload from the trainer.
func (h *ConfirmationHandler) handlePhoto(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	trainingID int64,
	state *entity.BotState,
) error {
	_ = answerCallback(h.bot, cb.ID, "Please upload a photo")

	telegramID := cb.From.ID
	chatID := cb.Message.Chat.ID

	var payload entity.TrainingConfirmPayload
	if state != nil {
		_ = json.Unmarshal([]byte(state.Payload), &payload)
	}
	payload.TrainingID = trainingID

	// Create or update the FSM state for photo upload.
	if state == nil {
		state = &entity.BotState{
			TelegramID: telegramID,
			ChatID:     chatID,
			Flow:       entity.FlowTrainingConfirm,
			Step:       1,
		}
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		stateID, err := h.botStateRepo.Create(ctx, state)
		if err != nil {
			return err
		}
		state.ID = stateID
	} else {
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		_ = h.botStateRepo.Update(ctx, state)
	}

	_ = sendText(h.bot, chatID, "Please send a photo for the training\\.")
	return nil
}

// HandlePhotoUpload processes a photo message when in the confirmation FSM.
func (h *ConfirmationHandler) HandlePhotoUpload(
	ctx context.Context,
	msg *tgbotapi.Message,
	state *entity.BotState,
) error {
	if len(msg.Photo) == 0 {
		_ = sendText(h.bot, state.ChatID, "That doesn't look like a photo\\. Please send an image\\.")
		return nil
	}

	// Get the largest photo size.
	photo := msg.Photo[len(msg.Photo)-1]

	var payload entity.TrainingConfirmPayload
	if err := json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return err
	}

	payload.PhotoFileID = photo.FileID
	payloadData, _ := json.Marshal(payload)
	state.Payload = string(payloadData)
	_ = h.botStateRepo.Update(ctx, state)

	_ = sendText(h.bot, state.ChatID, "Photo saved\\! You can now confirm the training\\.")

	// Show the confirmation keyboard again.
	keyboard := tgutil.ConfirmationKeyboard(payload.TrainingID)
	_ = sendInlineKeyboard(h.bot, state.ChatID, "Ready to confirm?", keyboard)
	return nil
}

// handleAddParticipant shows a member list to add to the training.
func (h *ConfirmationHandler) handleAddParticipant(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	trainingID int64,
	state *entity.BotState,
) error {
	_ = answerCallback(h.bot, cb.ID, "")

	training, err := h.trainingUC.GetTraining(ctx, trainingID)
	if err != nil {
		return err
	}

	members, err := h.memberUC.ListMembers(ctx, training.ClubID)
	if err != nil {
		return err
	}

	// Create or update the FSM state.
	telegramID := cb.From.ID
	chatID := cb.Message.Chat.ID

	var payload entity.TrainingConfirmPayload
	if state != nil {
		_ = json.Unmarshal([]byte(state.Payload), &payload)
	}
	payload.TrainingID = trainingID

	if state == nil {
		state = &entity.BotState{
			TelegramID: telegramID,
			ChatID:     chatID,
			Flow:       entity.FlowTrainingConfirm,
			Step:       2,
		}
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		var stateID int64
		stateID, err = h.botStateRepo.Create(ctx, state)
		if err != nil {
			return err
		}
		state.ID = stateID
	} else {
		state.Step = 2
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		_ = h.botStateRepo.Update(ctx, state)
	}

	// Show member selection keyboard with "addsel:training:id:memberID" format.
	keyboard := tgutil.MemberSelectKeyboard(members, ActionAddSelect, trainingID)
	_ = sendInlineKeyboard(h.bot, chatID, "Select a member to add:", keyboard)
	return nil
}

// handleRemoveParticipant shows the participant list to remove from the training.
func (h *ConfirmationHandler) handleRemoveParticipant(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	trainingID int64,
	state *entity.BotState,
) error {
	_ = answerCallback(h.bot, cb.ID, "")

	participants, err := h.trainingUC.ListParticipants(ctx, trainingID)
	if err != nil {
		return err
	}

	telegramID := cb.From.ID
	chatID := cb.Message.Chat.ID

	var payload entity.TrainingConfirmPayload
	if state != nil {
		_ = json.Unmarshal([]byte(state.Payload), &payload)
	}
	payload.TrainingID = trainingID

	if state == nil {
		state = &entity.BotState{
			TelegramID: telegramID,
			ChatID:     chatID,
			Flow:       entity.FlowTrainingConfirm,
			Step:       3,
		}
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		var stateID int64
		stateID, err = h.botStateRepo.Create(ctx, state)
		if err != nil {
			return err
		}
		state.ID = stateID
	} else {
		state.Step = 3
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		_ = h.botStateRepo.Update(ctx, state)
	}

	// Show participant selection keyboard with "remsel:training:id:memberID" format.
	keyboard := tgutil.MemberSelectKeyboard(participants, ActionRemSelect, trainingID)
	_ = sendInlineKeyboard(h.bot, chatID, "Select a participant to remove:", keyboard)
	return nil
}

// HandleAddParticipantCallback handles the selection of a member to add during confirmation.
func (h *ConfirmationHandler) HandleAddParticipantCallback(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	trainingID int64,
	memberID int64,
) error {
	_ = answerCallback(h.bot, cb.ID, "Member added")

	if err := h.trainingUC.AddParticipant(ctx, trainingID, memberID); err != nil {
		return fmt.Errorf("add participant: %w", err)
	}

	chatID := cb.Message.Chat.ID
	h.updateConfirmPayload(ctx, cb.From.ID, chatID, func(p *entity.TrainingConfirmPayload) {
		p.AddedIDs = append(p.AddedIDs, memberID)
	})

	_ = sendText(h.bot, chatID, "Participant added\\.")

	keyboard := tgutil.ConfirmationKeyboard(trainingID)
	_ = sendInlineKeyboard(h.bot, chatID, "Continue with confirmation?", keyboard)
	return nil
}

// HandleRemoveParticipantCallback handles the selection of a participant to remove during confirmation.
func (h *ConfirmationHandler) HandleRemoveParticipantCallback(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	trainingID int64,
	memberID int64,
) error {
	_ = answerCallback(h.bot, cb.ID, "Participant removed")

	if err := h.trainingUC.RemoveParticipant(ctx, trainingID, memberID); err != nil {
		return fmt.Errorf("remove participant: %w", err)
	}

	chatID := cb.Message.Chat.ID
	h.updateConfirmPayload(ctx, cb.From.ID, chatID, func(p *entity.TrainingConfirmPayload) {
		p.RemovedIDs = append(p.RemovedIDs, memberID)
	})

	_ = sendText(h.bot, chatID, "Participant removed\\.")

	keyboard := tgutil.ConfirmationKeyboard(trainingID)
	_ = sendInlineKeyboard(h.bot, chatID, "Continue with confirmation?", keyboard)
	return nil
}

// updateConfirmPayload updates the confirmation FSM payload by applying the given modifier function.
func (h *ConfirmationHandler) updateConfirmPayload(
	ctx context.Context,
	telegramID int64,
	chatID int64,
	modify func(*entity.TrainingConfirmPayload),
) {
	state, _ := h.botStateRepo.GetByTelegramAndFlow(ctx, telegramID, chatID, entity.FlowTrainingConfirm)
	if state == nil {
		return
	}
	var payload entity.TrainingConfirmPayload
	if err := json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return
	}
	modify(&payload)
	payloadData, _ := json.Marshal(payload)
	state.Payload = string(payloadData)
	_ = h.botStateRepo.Update(ctx, state)
}
