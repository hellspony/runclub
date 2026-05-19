package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"runclub/internal/delivery/telegram/tgutil"
	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
	"runclub/internal/pkg/templater"
	"runclub/internal/pkg/tgrescape"
	"runclub/internal/usecase"
)

// TrainingHandler handles the training creation flow (7-step FSM).
type TrainingHandler struct {
	bot          *tgbotapi.BotAPI
	clubUC       usecase.ClubUseCase
	locationUC   usecase.LocationUseCase
	trainingUC   usecase.TrainingUseCase
	memberUC     usecase.MemberUseCase
	templateUC   usecase.TemplateUseCase
	botStateRepo repository.BotStateRepository
	logger       *zap.Logger
}

// NewTrainingHandler creates a new TrainingHandler.
func NewTrainingHandler(
	bot *tgbotapi.BotAPI,
	clubUC usecase.ClubUseCase,
	locationUC usecase.LocationUseCase,
	trainingUC usecase.TrainingUseCase,
	memberUC usecase.MemberUseCase,
	templateUC usecase.TemplateUseCase,
	botStateRepo repository.BotStateRepository,
	logger *zap.Logger,
) *TrainingHandler {
	return &TrainingHandler{
		bot:          bot,
		clubUC:       clubUC,
		locationUC:   locationUC,
		trainingUC:   trainingUC,
		memberUC:     memberUC,
		templateUC:   templateUC,
		botStateRepo: botStateRepo,
		logger:       logger,
	}
}

// HandleCommand starts the training creation flow when /training is sent.
func (h *TrainingHandler) HandleCommand(ctx context.Context, msg *tgbotapi.Message) error {
	telegramID := msg.From.ID
	chatID := msg.Chat.ID

	// Check for an existing active state.
	existing, err := h.botStateRepo.GetByTelegramAndFlow(ctx, telegramID, chatID, entity.FlowTrainingCreate)
	if err == nil && existing != nil {
		_ = sendText(h.bot, chatID, "You already have an active training creation flow\\. Please complete it first\\.")
		return nil
	}

	// Get the member and their clubs.
	member, err := h.memberUC.GetByTelegramID(ctx, telegramID)
	if err != nil {
		_ = sendText(h.bot, chatID, "You are not registered\\. Please join a club first\\.")
		return nil //nolint:nilerr // intentional: user feedback already sent
	}

	trainerClubs, err := h.memberUC.ListTrainerClubs(ctx, member.ID)
	if err != nil || len(trainerClubs) == 0 {
		_ = sendText(h.bot, chatID, "You are not a trainer in any club\\.")
		return nil //nolint:nilerr // intentional: user feedback already sent
	}

	// Get club details for the keyboard.
	var clubs []entity.Club
	for _, tc := range trainerClubs {
		club, clubErr := h.clubUC.GetByID(ctx, tc.ClubID)
		if clubErr != nil {
			continue
		}
		clubs = append(clubs, *club)
	}

	if len(clubs) == 0 {
		_ = sendText(h.bot, chatID, "No clubs found\\.")
		return nil
	}

	// Create the initial bot state.
	state := &entity.BotState{
		TelegramID: telegramID,
		ChatID:     chatID,
		Flow:       entity.FlowTrainingCreate,
		Step:       1,
	}
	stateID, err := h.botStateRepo.Create(ctx, state)
	if err != nil {
		return fmt.Errorf("create bot state: %w", err)
	}
	state.ID = stateID

	// Initialize payload.
	payload := &entity.TrainingCreatePayload{}
	payloadData, _ := json.Marshal(payload)
	state.Payload = string(payloadData)
	_ = h.botStateRepo.Update(ctx, state)

	// Send club selection keyboard.
	keyboard := tgutil.ClubKeyboardFromClubs(clubs, ActionSelect)
	_ = sendInlineKeyboard(h.bot, chatID, "Select a club:", keyboard)

	return nil
}

// HandleStep processes a step in the training creation FSM.
func (h *TrainingHandler) HandleStep(ctx context.Context, msg *tgbotapi.Message, state *entity.BotState) error {
	switch state.Step {
	case 3:
		return h.handleNewLocationInput(ctx, msg, state)
	case 4:
		return h.handleDateTimeInput(ctx, msg, state)
	case 5:
		return h.handleDurationInput(ctx, msg, state)
	default:
		h.logger.Warn("unexpected step in training creation FSM",
			zap.Int("step", state.Step),
		)
		return nil
	}
}

// HandleCallback processes callback queries for the training flow.
func (h *TrainingHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	data := tgutil.Decode(cb.Data)
	if len(data) < 3 {
		return nil
	}

	action := data[0]
	entityType := data[1]

	_ = answerCallback(h.bot, cb.ID, "")

	chatID := cb.Message.Chat.ID
	telegramID := cb.From.ID

	state, err := h.botStateRepo.GetByTelegramAndFlow(ctx, telegramID, chatID, entity.FlowTrainingCreate)
	if err != nil || state == nil {
		return nil //nolint:nilerr // intentional: user feedback already sent
	}

	var payload entity.TrainingCreatePayload
	if err = json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	switch {
	case action == ActionSelect && entityType == EntityClub && state.Step == 1:
		return h.handleClubSelect(ctx, state, &payload, data[2])
	case action == ActionSelect && entityType == EntityLocation && state.Step == 2:
		return h.handleLocationSelect(ctx, state, &payload, data[2])
	case action == ActionNew && entityType == EntityLocation && state.Step == 2:
		return h.handleNewLocationStep(ctx, state, &payload, data[2])
	case (action == ActionAddSelect || action == ActionRemSelect) && entityType == EntityTrainer && state.Step == 6:
		return h.handleTrainerToggle(ctx, state, &payload, data[2])
	case action == ActionDone && entityType == EntityTrainer && state.Step == 6:
		return h.createTraining(ctx, state, &payload)
	default:
		return nil
	}
}

func (h *TrainingHandler) handleClubSelect(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.TrainingCreatePayload,
	clubIDStr string,
) error {
	clubID, err := parseInt64(clubIDStr)
	if err != nil {
		return err
	}
	payload.ClubID = clubIDStr
	if err = h.savePayloadAndStep(ctx, state, payload, 2); err != nil {
		return err
	}

	locations, err := h.locationUC.ListByClub(ctx, clubID)
	if err != nil {
		return err
	}
	keyboard := tgutil.LocationKeyboard(locations, clubID)
	_ = sendInlineKeyboard(h.bot, state.ChatID, "Select a location:", keyboard)
	return nil
}

func (h *TrainingHandler) handleLocationSelect(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.TrainingCreatePayload,
	locationIDStr string,
) error {
	payload.LocationID = locationIDStr
	if err := h.savePayloadAndStep(ctx, state, payload, 4); err != nil {
		return err
	}
	_ = sendText(h.bot, state.ChatID,
		"Enter date and time \\(e\\.g\\. 2024\\-01\\-15 18:00 or 15\\.01\\.2024 18:00\\):")
	return nil
}

func (h *TrainingHandler) handleNewLocationStep(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.TrainingCreatePayload,
	clubIDStr string,
) error {
	payload.ClubID = clubIDStr
	state.Step = 3
	if err := h.savePayloadAndStep(ctx, state, payload, 3); err != nil {
		return err
	}
	_ = sendText(h.bot, state.ChatID, "Enter location details in format:\\n`Name | Address | Map URL | Description`")
	return nil
}

func (h *TrainingHandler) handleTrainerToggle(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.TrainingCreatePayload,
	trainerIDStr string,
) error {
	trainerID, err := parseInt64(trainerIDStr)
	if err != nil {
		return err
	}
	payload.TrainerIDs = toggleID(payload.TrainerIDs, trainerID)
	if err = h.savePayloadAndStep(ctx, state, payload, 6); err != nil {
		return err
	}

	clubID, _ := parseInt64(payload.ClubID)
	trainers, err := h.memberUC.ListTrainers(ctx, clubID)
	if err != nil {
		return err
	}
	keyboard := tgutil.TrainerKeyboard(trainers, payload.TrainerIDs, clubID)
	_ = sendInlineKeyboard(h.bot, state.ChatID, "Select trainers:", keyboard)
	return nil
}

// handleNewLocationInput processes text input for a new location (step 3).
func (h *TrainingHandler) handleNewLocationInput(
	ctx context.Context,
	msg *tgbotapi.Message,
	state *entity.BotState,
) error {
	var payload entity.TrainingCreatePayload
	if err := json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return err
	}

	// Parse the input: "Name | Address | Map URL | Description"
	parts := splitPipe(msg.Text)
	var name, address, mapURL, description string

	switch len(parts) {
	case 4:
		description = parts[3]
		fallthrough
	case 3:
		mapURL = parts[2]
		fallthrough
	case 2:
		address = parts[1]
		fallthrough
	case 1:
		name = parts[0]
	default:
		_ = sendText(h.bot, state.ChatID, "Invalid format\\. Please use:\\n`Name | Address | Map URL | Description`")
		return nil
	}

	clubID, _ := parseInt64(payload.ClubID)
	location := &entity.Location{
		ClubID:      clubID,
		Name:        name,
		Address:     address,
		MapURL:      mapURL,
		Description: description,
	}

	locID, err := h.locationUC.Create(ctx, location)
	if err != nil {
		return fmt.Errorf("create location: %w", err)
	}

	payload.LocationID = strconv.FormatInt(locID, 10)
	payloadData, _ := json.Marshal(payload)
	state.Payload = string(payloadData)
	state.Step = 4
	_ = h.botStateRepo.Update(ctx, state)

	_ = sendText(h.bot, state.ChatID, "Location created\\! Enter date and time \\(e\\.g\\. 2024\\-01\\-15 18:00\\):")
	return nil
}

// handleDateTimeInput processes date/time text input (step 4).
func (h *TrainingHandler) handleDateTimeInput(
	ctx context.Context,
	msg *tgbotapi.Message,
	state *entity.BotState,
) error {
	var payload entity.TrainingCreatePayload
	if err := json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return err
	}

	dt, err := parseDateTime(msg.Text)
	if err != nil {
		_ = sendText(
			h.bot,
			state.ChatID,
			"Cannot parse date/time\\. Try: `2024\\-01\\-15 18:00` or `15\\.01\\.2024 18:00`",
		)
		return nil //nolint:nilerr // intentional: user feedback already sent
	}

	payload.Date = dt.Format(timeFormat)
	payloadData, _ := json.Marshal(payload)
	state.Payload = string(payloadData)
	state.Step = 5
	_ = h.botStateRepo.Update(ctx, state)

	_ = sendText(h.bot, state.ChatID, "Enter duration in minutes:")
	return nil
}

// handleDurationInput processes duration text input (step 5).
func (h *TrainingHandler) handleDurationInput(
	ctx context.Context,
	msg *tgbotapi.Message,
	state *entity.BotState,
) error {
	var payload entity.TrainingCreatePayload
	if err := json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return err
	}

	duration, err := strconv.Atoi(msg.Text)
	if err != nil || duration <= 0 {
		_ = sendText(h.bot, state.ChatID, "Please enter a valid duration in minutes:")
		return nil //nolint:nilerr // intentional: user feedback already sent
	}

	payload.Duration = duration
	payloadData, _ := json.Marshal(payload)
	state.Payload = string(payloadData)
	state.Step = 6
	_ = h.botStateRepo.Update(ctx, state)

	// Show trainer selection keyboard.
	clubID, _ := parseInt64(payload.ClubID)
	trainers, err := h.memberUC.ListTrainers(ctx, clubID)
	if err != nil {
		return err
	}
	keyboard := tgutil.TrainerKeyboard(trainers, payload.TrainerIDs, clubID)
	_ = sendInlineKeyboard(h.bot, state.ChatID, "Select trainers:", keyboard)
	return nil
}

// createTraining is step 7: create DB records and post to the club chat.
func (h *TrainingHandler) createTraining(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.TrainingCreatePayload,
) error {
	clubID, _ := parseInt64(payload.ClubID)
	locationID, _ := parseInt64(payload.LocationID)
	dt, _ := time.Parse(timeFormat, payload.Date)

	training, err := h.trainingUC.CreateTraining(ctx, clubID, locationID, dt, payload.Duration, payload.TrainerIDs)
	if err != nil {
		return fmt.Errorf("create training: %w", err)
	}

	// Get club for chat ID.
	club, err := h.clubUC.GetByID(ctx, clubID)
	if err != nil {
		return err
	}

	// Get location.
	location, err := h.locationUC.GetByID(ctx, locationID)
	if err != nil {
		return err
	}

	// Get trainer names.
	var trainerNames []string
	for _, tid := range payload.TrainerIDs {
		m, mErr := h.memberUC.GetByID(ctx, tid)
		if mErr != nil {
			continue
		}
		trainerNames = append(trainerNames, tgutil.DisplayName(*m))
	}

	// Render template.
	tmpl, err := h.templateUC.GetByClubAndType(ctx, clubID, entity.TemplateTrainingNew)
	var text string
	if err != nil || tmpl == nil {
		// Fallback template.
		text = fmt.Sprintf("Training: %s, %s, %d min, Trainers: %v",
			tgrescape.EscapeMarkdownV2(location.Name),
			tgrescape.EscapeMarkdownV2(dt.Format("02.01.2006 15:04")),
			payload.Duration,
			trainerNames,
		)
	} else {
		text, err = templater.Render(tmpl.Content, map[string]any{
			TplLocationName: location.Name,
			TplDate:         dt.Format("02.01.2006 15:04"),
			"Duration":      payload.Duration,
			"Trainers":      trainerNames,
		})
		if err != nil {
			text = "New training scheduled!"
		}
		text = tgrescape.EscapeMarkdownV2(text)
	}

	// Post to club chat with participation keyboard.
	keyboard := tgutil.ParticipationKeyboard(training.ID, false)
	sent, err := sendInlineKeyboardWithResult(h.bot, club.TelegramChatID, text, keyboard)
	if err != nil {
		return fmt.Errorf("post training to chat: %w", err)
	}

	// Pin the message.
	pin := tgbotapi.PinChatMessageConfig{
		ChatID:    club.TelegramChatID,
		MessageID: sent.MessageID,
	}
	_, _ = h.bot.Request(pin)

	// Save message ID on the training record.
	training.MessageID = int64(sent.MessageID)
	_ = h.trainingUC.UpdateTraining(ctx, training)

	// Clear the FSM state.
	_ = h.botStateRepo.Delete(ctx, state.ID)

	_ = sendText(h.bot, state.ChatID, "Training created and posted\\!")
	return nil
}

const timeFormat = "2006-01-02 15:04:05 -0700 MST"

// savePayloadAndStep serializes the payload and updates the state step.
func (h *TrainingHandler) savePayloadAndStep(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.TrainingCreatePayload,
	step int,
) error {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	state.Payload = string(payloadData)
	state.Step = step
	return h.botStateRepo.Update(ctx, state)
}

// splitPipe splits a string by " | " or "|".
func splitPipe(s string) []string {
	var result []string
	current := ""
	for _, ch := range s {
		if ch == '|' {
			result = append(result, trimSpace(current))
			current = ""
		} else {
			current += string(ch)
		}
	}
	result = append(result, trimSpace(current))
	return result
}

func trimSpace(s string) string {
	for len(s) > 0 && s[0] == ' ' {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}

// sendInlineKeyboardWithResult sends a message with an inline keyboard and returns the sent message.
func sendInlineKeyboardWithResult(
	bot *tgbotapi.BotAPI,
	chatID int64,
	text string,
	keyboard tgbotapi.InlineKeyboardMarkup,
) (*tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = keyboard
	sent, err := bot.Send(msg)
	if err != nil {
		return nil, err
	}
	return &sent, nil
}
