package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"runclub/internal/delivery/telegram/tgutil"
	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
	"runclub/internal/pkg/templater"
	"runclub/internal/pkg/tgrescape"
	"runclub/internal/usecase"
)

// JointRunHandler handles the joint run creation flow (4-step FSM).
type JointRunHandler struct {
	bot          *tgbotapi.BotAPI
	clubUC       usecase.ClubUseCase
	locationUC   usecase.LocationUseCase
	jointRunUC   usecase.JointRunUseCase
	memberUC     usecase.MemberUseCase
	templateUC   usecase.TemplateUseCase
	botStateRepo repository.BotStateRepository
	logger       *zap.Logger
}

// NewJointRunHandler creates a new JointRunHandler.
func NewJointRunHandler(
	bot *tgbotapi.BotAPI,
	clubUC usecase.ClubUseCase,
	locationUC usecase.LocationUseCase,
	jointRunUC usecase.JointRunUseCase,
	memberUC usecase.MemberUseCase,
	templateUC usecase.TemplateUseCase,
	botStateRepo repository.BotStateRepository,
	logger *zap.Logger,
) *JointRunHandler {
	return &JointRunHandler{
		bot:          bot,
		clubUC:       clubUC,
		locationUC:   locationUC,
		jointRunUC:   jointRunUC,
		memberUC:     memberUC,
		templateUC:   templateUC,
		botStateRepo: botStateRepo,
		logger:       logger,
	}
}

// HandleCommand starts the joint run creation flow when /jointrun is sent.
func (h *JointRunHandler) HandleCommand(ctx context.Context, msg *tgbotapi.Message) error {
	telegramID := msg.From.ID
	chatID := msg.Chat.ID

	// Check for an existing active state.
	existing, err := h.botStateRepo.GetByTelegramAndFlow(ctx, telegramID, chatID, entity.FlowJointRunCreate)
	if err == nil && existing != nil {
		_ = sendText(h.bot, chatID, "You already have an active joint run creation flow\\. Please complete it first\\.")
		return nil
	}

	// Get the member and their clubs.
	member, err := h.memberUC.GetByTelegramID(ctx, telegramID)
	if err != nil {
		_ = sendText(h.bot, chatID, "You are not registered\\. Please join a club first\\.")
		return nil //nolint:nilerr // intentional: user feedback already sent
	}

	clubsMember, err := h.memberUC.ListTrainerClubs(ctx, member.ID)
	if err != nil || len(clubsMember) == 0 {
		// Try regular clubs as well.
		clubsMember, err = h.memberUC.ListTrainerClubs(ctx, member.ID)
		if err != nil || len(clubsMember) == 0 {
			_ = sendText(h.bot, chatID, "You are not a member of any club\\.")
			return nil //nolint:nilerr // intentional: user feedback already sent
		}
	}

	// Get club details for the keyboard.
	var clubs []entity.Club
	for _, tc := range clubsMember {
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
		Flow:       entity.FlowJointRunCreate,
		Step:       1,
	}
	stateID, err := h.botStateRepo.Create(ctx, state)
	if err != nil {
		return fmt.Errorf("create bot state: %w", err)
	}
	state.ID = stateID

	// Initialize payload.
	payload := &entity.JointRunCreatePayload{}
	payloadData, _ := json.Marshal(payload)
	state.Payload = string(payloadData)
	_ = h.botStateRepo.Update(ctx, state)

	// Send club selection keyboard.
	keyboard := tgutil.ClubKeyboardFromClubs(clubs, ActionSelect)
	_ = sendInlineKeyboard(h.bot, chatID, "Select a club:", keyboard)

	return nil
}

// HandleStep processes a step in the joint run creation FSM.
func (h *JointRunHandler) HandleStep(ctx context.Context, msg *tgbotapi.Message, state *entity.BotState) error {
	switch state.Step {
	case 3:
		return h.handleNewLocationInput(ctx, msg, state)
	case 4:
		return h.handleDateTimeInput(ctx, msg, state)
	default:
		h.logger.Warn("unexpected step in joint run creation FSM",
			zap.Int("step", state.Step),
		)
		return nil
	}
}

// HandleCallback processes callback queries for the joint run flow.
func (h *JointRunHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	data := tgutil.Decode(cb.Data)
	if len(data) < 3 {
		return nil
	}

	action := data[0]
	entityType := data[1]

	_ = answerCallback(h.bot, cb.ID, "")

	chatID := cb.Message.Chat.ID
	telegramID := cb.From.ID

	state, err := h.botStateRepo.GetByTelegramAndFlow(ctx, telegramID, chatID, entity.FlowJointRunCreate)
	if err != nil || state == nil {
		return nil //nolint:nilerr // intentional: user feedback already sent
	}

	var payload entity.JointRunCreatePayload
	if err = json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	switch {
	case action == ActionSelect && entityType == EntityClub && state.Step == 1:
		var clubID int64
		clubID, err = parseInt64(data[2])
		if err != nil {
			return err
		}
		payload.ClubID = data[2]
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		state.Step = 2
		_ = h.botStateRepo.Update(ctx, state)

		// Show location selection.
		var locations []entity.Location
		locations, err = h.locationUC.ListByClub(ctx, clubID)
		if err != nil {
			return err
		}
		keyboard := tgutil.LocationKeyboard(locations, clubID)
		_ = sendInlineKeyboard(h.bot, chatID, "Select a location:", keyboard)
		return nil

	case action == ActionSelect && entityType == EntityLocation && state.Step == 2:
		payload.LocationID = data[2]
		// Move to step 4 (date/time) - skip step 3.
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		state.Step = 4
		_ = h.botStateRepo.Update(ctx, state)
		_ = sendText(h.bot, chatID, "Enter date and time \\(e\\.g\\. 2024\\-01\\-15 18:00 or 15\\.01\\.2024 18:00\\):")
		return nil

	case action == ActionNew && entityType == EntityLocation && state.Step == 2:
		// Move to step 3 (new location creation).
		state.Step = 3
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)
		_ = h.botStateRepo.Update(ctx, state)
		_ = sendText(h.bot, chatID, "Enter location details in format:\\n`Name | Address | Map URL | Description`")
		return nil

	default:
		return nil
	}
}

// handleNewLocationInput processes text input for a new location (step 3).
func (h *JointRunHandler) handleNewLocationInput(
	ctx context.Context,
	msg *tgbotapi.Message,
	state *entity.BotState,
) error {
	var payload entity.JointRunCreatePayload
	if err := json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return err
	}

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

	payload.LocationID = formatInt64(locID)
	payloadData, _ := json.Marshal(payload)
	state.Payload = string(payloadData)
	state.Step = 4
	_ = h.botStateRepo.Update(ctx, state)

	_ = sendText(h.bot, state.ChatID, "Location created\\! Enter date and time \\(e\\.g\\. 2024\\-01\\-15 18:00\\):")
	return nil
}

// handleDateTimeInput processes date/time text input (step 4) and creates the joint run.
func (h *JointRunHandler) handleDateTimeInput(
	ctx context.Context,
	msg *tgbotapi.Message,
	state *entity.BotState,
) error {
	var payload entity.JointRunCreatePayload
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

	clubID, _ := parseInt64(payload.ClubID)
	locationID, _ := parseInt64(payload.LocationID)

	// Get the member.
	member, err := h.memberUC.GetByTelegramID(ctx, state.TelegramID)
	if err != nil {
		return err
	}

	// Create the joint run.
	run, err := h.jointRunUC.CreateJointRun(ctx, clubID, locationID, member.ID, dt)
	if err != nil {
		return fmt.Errorf("create joint run: %w", err)
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

	// Render template.
	tmpl, err := h.templateUC.GetByClubAndType(ctx, clubID, entity.TemplateJointRunNew)
	var text string
	if err != nil || tmpl == nil {
		text = fmt.Sprintf("Joint Run: %s, %s",
			tgrescape.EscapeMarkdownV2(location.Name),
			tgrescape.EscapeMarkdownV2(dt.Format("02.01.2006 15:04")),
		)
	} else {
		text, err = templater.Render(tmpl.Content, map[string]any{
			TplLocationName: location.Name,
			TplDate:         dt.Format("02.01.2006 15:04"),
			"Creator":       tgutil.DisplayName(*member),
		})
		if err != nil {
			text = "New joint run scheduled!"
		}
		text = tgrescape.EscapeMarkdownV2(text)
	}

	// Post to club chat with participation keyboard.
	keyboard := tgutil.ParticipationKeyboardJointRun(run.ID, false)
	sent, err := sendInlineKeyboardWithResult(h.bot, club.TelegramChatID, text, keyboard)
	if err != nil {
		return fmt.Errorf("post joint run to chat: %w", err)
	}

	// Save message ID on the joint run record.
	run.MessageID = int64(sent.MessageID)
	_ = h.jointRunUC.UpdateJointRun(ctx, run)

	// Clear the FSM state.
	_ = h.botStateRepo.Delete(ctx, state.ID)

	_ = sendText(h.bot, state.ChatID, "Joint run created and posted\\!")
	return nil
}
