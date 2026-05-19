package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

// WelcomeHandler handles new chat member events and the welcome collection flow.
type WelcomeHandler struct {
	bot                *tgbotapi.BotAPI
	clubUC             usecase.ClubUseCase
	memberUC           usecase.MemberUseCase
	templateUC         usecase.TemplateUseCase
	customFieldRepo    repository.CustomFieldRepository
	customFieldValRepo repository.CustomFieldValueRepository
	botStateRepo       repository.BotStateRepository
	logger             *zap.Logger
}

// NewWelcomeHandler creates a new WelcomeHandler.
func NewWelcomeHandler(
	bot *tgbotapi.BotAPI,
	clubUC usecase.ClubUseCase,
	memberUC usecase.MemberUseCase,
	templateUC usecase.TemplateUseCase,
	customFieldRepo repository.CustomFieldRepository,
	customFieldValRepo repository.CustomFieldValueRepository,
	botStateRepo repository.BotStateRepository,
	logger *zap.Logger,
) *WelcomeHandler {
	return &WelcomeHandler{
		bot:                bot,
		clubUC:             clubUC,
		memberUC:           memberUC,
		templateUC:         templateUC,
		customFieldRepo:    customFieldRepo,
		customFieldValRepo: customFieldValRepo,
		botStateRepo:       botStateRepo,
		logger:             logger,
	}
}

// HandleChatMemberUpdate processes ChatMember updates when a user joins or leaves a group.
func (h *WelcomeHandler) HandleChatMemberUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.ChatMember == nil {
		return nil
	}

	chatMember := update.ChatMember
	newStatus := chatMember.NewChatMember.Status
	oldStatus := chatMember.OldChatMember.Status

	chatID := chatMember.Chat.ID
	user := chatMember.NewChatMember.User

	// Handle user leaving the chat
	if newStatus == "left" || newStatus == "kicked" {
		return h.handleMemberLeft(ctx, chatID, user.ID)
	}

	// Handle user joining the chat
	if newStatus != "member" && newStatus != "administrator" {
		return nil
	}
	if oldStatus == "member" || oldStatus == "administrator" {
		return nil
	}

	return h.handleMemberJoined(ctx, chatID, user)
}

func (h *WelcomeHandler) handleMemberLeft(ctx context.Context, chatID, telegramID int64) error {
	club, err := h.clubUC.GetByTelegramChatID(ctx, chatID)
	if err != nil {
		h.logger.Debug("club not found for chat on member left", zap.Int64("chat_id", chatID), zap.Error(err))
		return nil
	}

	member, err := h.memberUC.GetByTelegramID(ctx, telegramID)
	if err != nil {
		h.logger.Debug("member not found on member left", zap.Int64("telegram_id", telegramID), zap.Error(err))
		return nil
	}

	return h.memberUC.RemoveFromClub(ctx, club.ID, member.ID)
}

func (h *WelcomeHandler) handleMemberJoined(ctx context.Context, chatID int64, user *tgbotapi.User) error {
	club, err := h.clubUC.GetByTelegramChatID(ctx, chatID)
	if err != nil {
		h.logger.Debug("club not found for chat", zap.Int64("chat_id", chatID), zap.Error(err))
		return nil
	}

	if !club.WelcomeEnabled {
		return nil
	}

	member, err := h.memberUC.RegisterOrGet(ctx, user.ID, user.UserName)
	if err != nil {
		return fmt.Errorf("register or get member: %w", err)
	}

	// Check if already a member of this club (returning user)
	existingCM, err := h.memberUC.GetClubMember(ctx, club.ID, member.ID)
	if err == nil && existingCM != nil {
		// Returning member — just clear left_at, no welcome message
		if member.LeftAt != nil {
			member.LeftAt = nil
			_ = h.memberUC.UpdateProfile(ctx, member.ID, member.FIO, member.TelegramUsername, member.BirthDate)
		}
		return nil
	}

	// New member — add to club and send welcome
	if err = h.memberUC.AddToClub(ctx, club.ID, member.ID, entity.RoleMember); err != nil {
		return fmt.Errorf("add to club: %w", err)
	}

	h.sendWelcomeAndDM(ctx, club, member, user)

	return nil
}

// sendWelcomeAndDM sends the welcome message to the chat and a DM to the user.
func (h *WelcomeHandler) sendWelcomeAndDM(
	ctx context.Context,
	club *entity.Club,
	member *entity.Member,
	user *tgbotapi.User,
) {
	tmpl, err := h.templateUC.GetByClubAndType(ctx, club.ID, entity.TemplateWelcome)
	var welcomeText string
	if err != nil || tmpl == nil {
		welcomeText = fmt.Sprintf("Welcome, %s!", tgrescape.EscapeMarkdownV2(tgutil.DisplayName(*member)))
	} else {
		welcomeText, err = templater.Render(tmpl.Content, map[string]any{
			"Name":     tgutil.DisplayName(*member),
			"Username": user.UserName,
			"ClubName": club.Name,
		})
		if err != nil {
			welcomeText = "Welcome!"
		}
		welcomeText = tgrescape.EscapeMarkdownV2(welcomeText)
	}

	_ = sendText(h.bot, club.TelegramChatID, welcomeText)

	dmChat := tgbotapi.NewMessage(user.ID, "Would you like to share your information with the club?")
	dmChat.ReplyMarkup = tgutil.WelcomeKeyboard()
	if _, err = h.bot.Send(dmChat); err != nil {
		h.logger.Warn("failed to send welcome DM",
			zap.Int64("telegram_id", user.ID),
			zap.Error(err),
		)
	}
}

// HandleWelcomeCallback processes wel:yes and wel:no callbacks.
func (h *WelcomeHandler) HandleWelcomeCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	data := tgutil.Decode(cb.Data)
	if len(data) < 2 || data[0] != ActionWelcome {
		return nil
	}

	_ = answerCallback(h.bot, cb.ID, "")

	telegramID := cb.From.ID
	chatID := cb.Message.Chat.ID
	choice := data[1] // "yes" or "no"

	member, err := h.memberUC.RegisterOrGet(ctx, telegramID, cb.From.UserName)
	if err != nil {
		return err
	}

	switch choice {
	case "no":
		// Just save telegram username + club relation. Clear any existing state.
		state, _ := h.botStateRepo.GetByTelegramAndFlow(ctx, telegramID, chatID, entity.FlowWelcomeCollect)
		if state != nil {
			_ = h.botStateRepo.Delete(ctx, state.ID)
		}
		_ = sendText(h.bot, chatID, "No problem\\! You can always share your info later\\.")
		return nil

	case "yes":
		// Start the welcome collection flow.
		// Check for existing state.
		existing, _ := h.botStateRepo.GetByTelegramAndFlow(ctx, telegramID, chatID, entity.FlowWelcomeCollect)
		if existing != nil {
			_ = sendText(h.bot, chatID, "You already have an active profile collection\\. Please complete it first\\.")
			return nil
		}

		// Find the club - we need to determine which club this is for.
		// The DM doesn't have a club chat ID, so we find the member's club.
		clubs, clubsErr := h.memberUC.ListClubsByMember(ctx, member.ID)
		if clubsErr != nil {
			h.logger.Debug("list clubs by member failed", zap.Int64("member_id", member.ID), zap.Error(clubsErr))
		}
		if clubsErr != nil || len(clubs) == 0 {
			_ = sendText(h.bot, chatID, "Could not find your club\\. Please join a club group first\\.")
			return nil //nolint:nilerr // intentional: user feedback already sent
		}

		clubID := clubs[0].ClubID

		state := &entity.BotState{
			TelegramID: telegramID,
			ChatID:     chatID,
			Flow:       entity.FlowWelcomeCollect,
			Step:       1,
		}

		payload := &entity.WelcomeCollectPayload{
			ClubID:   formatInt64(clubID),
			MemberID: formatInt64(member.ID),
		}
		payloadData, _ := json.Marshal(payload)
		state.Payload = string(payloadData)

		var stateID int64
		stateID, err = h.botStateRepo.Create(ctx, state)
		if err != nil {
			return err
		}
		state.ID = stateID

		_ = sendText(h.bot, chatID, "Please enter your full name \\(FIO\\):")
		return nil

	default:
		return nil
	}
}

// HandleStep processes a step in the welcome collection FSM.
func (h *WelcomeHandler) HandleStep(ctx context.Context, msg *tgbotapi.Message, state *entity.BotState) error {
	var payload entity.WelcomeCollectPayload
	if err := json.Unmarshal([]byte(state.Payload), &payload); err != nil {
		return err
	}

	switch state.Step {
	case 1:
		return h.handleFIOStep(ctx, state, &payload, msg.Text)
	case 2:
		return h.handleBirthDateStep(ctx, state, &payload, msg.Text)
	case 3:
		return h.handleCustomFieldStep(ctx, state, &payload, msg.Text)
	default:
		return nil
	}
}

func (h *WelcomeHandler) handleFIOStep(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.WelcomeCollectPayload,
	text string,
) error {
	payload.FIO = text
	if err := h.savePayloadAndStep(ctx, state, payload, 2); err != nil {
		return err
	}
	_ = sendText(h.bot, state.ChatID, "Enter your birth date \\(DD\\.MM\\.YYYY\\):")
	return nil
}

func (h *WelcomeHandler) handleBirthDateStep(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.WelcomeCollectPayload,
	text string,
) error {
	birthDate, err := time.Parse("02.01.2006", text)
	if err != nil {
		_ = sendText(h.bot, state.ChatID, "Cannot parse date\\. Use DD\\.MM\\.YYYY format:")
		return nil //nolint:nilerr // intentional: user feedback already sent
	}
	payload.BirthDate = birthDate.Format("2006-01-02")

	clubID, _ := parseInt64(payload.ClubID)
	customFields, err := h.customFieldRepo.ListByClub(ctx, clubID)
	if err != nil || len(customFields) == 0 {
		return h.finishWelcomeFlow(ctx, state, payload)
	}

	if payload.CustomFields == nil {
		payload.CustomFields = make(map[int64]string)
	}
	payload.CurrentField = customFields[0].ID
	if err = h.savePayloadAndStep(ctx, state, payload, 3); err != nil {
		return err
	}

	_ = sendText(h.bot, state.ChatID, tgrescape.EscapeMarkdownV2(customFields[0].Name)+":")
	return nil
}

func (h *WelcomeHandler) handleCustomFieldStep(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.WelcomeCollectPayload,
	text string,
) error {
	if payload.CustomFields == nil {
		payload.CustomFields = make(map[int64]string)
	}
	payload.CustomFields[payload.CurrentField] = text

	clubID, _ := parseInt64(payload.ClubID)
	customFields, err := h.customFieldRepo.ListByClub(ctx, clubID)
	if err != nil {
		return err
	}

	nextField := h.findNextCustomField(customFields, payload.CurrentField)
	if nextField != nil {
		payload.CurrentField = nextField.ID
		if err = h.savePayloadAndStep(ctx, state, payload, 3); err != nil {
			return err
		}
		_ = sendText(h.bot, state.ChatID, tgrescape.EscapeMarkdownV2(nextField.Name)+":")
		return nil
	}

	return h.finishWelcomeFlow(ctx, state, payload)
}

// findNextCustomField returns the field after the one with the given ID, or nil if none.
func (h *WelcomeHandler) findNextCustomField(fields []entity.CustomField, currentID int64) *entity.CustomField {
	foundCurrent := false
	for i := range fields {
		if foundCurrent {
			return &fields[i]
		}
		if fields[i].ID == currentID {
			foundCurrent = true
		}
	}
	return nil
}

// finishWelcomeFlow saves all collected data and clears the FSM state.
func (h *WelcomeHandler) finishWelcomeFlow(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.WelcomeCollectPayload,
) error {
	memberID, _ := parseInt64(payload.MemberID)

	// Update member profile.
	var birthDate *time.Time
	if payload.BirthDate != "" {
		bd, err := time.Parse("2006-01-02", payload.BirthDate)
		if err == nil {
			birthDate = &bd
		}
	}
	if err := h.memberUC.UpdateProfile(ctx, memberID, payload.FIO, "", birthDate); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	// Save custom field values.
	for fieldID, value := range payload.CustomFields {
		cfv := &entity.CustomFieldValue{
			MemberID:      memberID,
			CustomFieldID: fieldID,
			Value:         value,
		}
		if err := h.customFieldValRepo.CreateOrUpdate(ctx, cfv); err != nil {
			h.logger.Error("failed to save custom field value",
				zap.Int64("field_id", fieldID),
				zap.Error(err),
			)
		}
	}

	// Clear FSM state.
	_ = h.botStateRepo.Delete(ctx, state.ID)

	_ = sendText(h.bot, state.ChatID, "Thank you\\! Your profile has been saved\\.")
	return nil
}

// savePayloadAndStep serializes the payload and updates the state step.
func (h *WelcomeHandler) savePayloadAndStep(
	ctx context.Context,
	state *entity.BotState,
	payload *entity.WelcomeCollectPayload,
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
