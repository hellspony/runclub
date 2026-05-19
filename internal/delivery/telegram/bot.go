package telegram

import (
	"context"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"runclub/internal/delivery/telegram/handlers"
	"runclub/internal/delivery/telegram/tgutil"
	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
	"runclub/internal/pkg/tgrescape"
	"runclub/internal/usecase"
)

// Bot is the Telegram bot delivery layer.
type Bot struct {
	api            *tgbotapi.BotAPI
	trainingH      *handlers.TrainingHandler
	jointrunH      *handlers.JointRunHandler
	participationH *handlers.ParticipationHandler
	welcomeH       *handlers.WelcomeHandler
	confirmationH  *handlers.ConfirmationHandler
	racePollH      *handlers.RacePollHandler
	router         *Router
	stateMachine   *StateMachine
	logger         *zap.Logger
	stopCh         chan struct{}
}

// NewBot creates a new Bot with all handlers wired up.
func NewBot(
	api *tgbotapi.BotAPI,
	clubUC usecase.ClubUseCase,
	locationUC usecase.LocationUseCase,
	trainingUC usecase.TrainingUseCase,
	jointRunUC usecase.JointRunUseCase,
	memberUC usecase.MemberUseCase,
	raceUC usecase.RaceUseCase,
	templateUC usecase.TemplateUseCase,
	customFieldRepo repository.CustomFieldRepository,
	customFieldValRepo repository.CustomFieldValueRepository,
	botStateRepo repository.BotStateRepository,
	logger *zap.Logger,
) *Bot {
	// Create handlers.
	trainingH := handlers.NewTrainingHandler(
		api, clubUC, locationUC, trainingUC, memberUC, templateUC, botStateRepo, logger,
	)
	jointrunH := handlers.NewJointRunHandler(
		api, clubUC, locationUC, jointRunUC, memberUC, templateUC, botStateRepo, logger,
	)
	participationH := handlers.NewParticipationHandler(
		api, trainingUC, jointRunUC, memberUC, clubUC, logger,
	)
	welcomeH := handlers.NewWelcomeHandler(
		api, clubUC, memberUC, templateUC, customFieldRepo, customFieldValRepo, botStateRepo, logger,
	)
	confirmationH := handlers.NewConfirmationHandler(
		api, trainingUC, memberUC, clubUC, templateUC, locationUC, botStateRepo, logger,
	)
	racePollH := handlers.NewRacePollHandler(
		api, raceUC, memberUC, logger,
	)

	// Create state machine.
	sm := NewStateMachine(botStateRepo, logger)

	// Register FSM steps.
	// Training creation steps (3, 4, 5 are text-input steps; 1, 2, 6, 7 are callback steps).
	sm.RegisterStep(entity.FlowTrainingCreate, 3, trainingH.HandleStep)
	sm.RegisterStep(entity.FlowTrainingCreate, 4, trainingH.HandleStep)
	sm.RegisterStep(entity.FlowTrainingCreate, 5, trainingH.HandleStep)

	// Joint run creation steps (3, 4 are text-input steps).
	sm.RegisterStep(entity.FlowJointRunCreate, 3, jointrunH.HandleStep)
	sm.RegisterStep(entity.FlowJointRunCreate, 4, jointrunH.HandleStep)

	// Welcome collection steps.
	sm.RegisterStep(entity.FlowWelcomeCollect, 1, welcomeH.HandleStep)
	sm.RegisterStep(entity.FlowWelcomeCollect, 2, welcomeH.HandleStep)
	sm.RegisterStep(entity.FlowWelcomeCollect, 3, welcomeH.HandleStep)

	// Training confirmation steps.
	sm.RegisterStep(entity.FlowTrainingConfirm, 1, confirmationH.HandlePhotoUpload)

	bot := &Bot{
		api:            api,
		trainingH:      trainingH,
		jointrunH:      jointrunH,
		participationH: participationH,
		welcomeH:       welcomeH,
		confirmationH:  confirmationH,
		racePollH:      racePollH,
		stateMachine:   sm,
		logger:         logger,
		stopCh:         make(chan struct{}),
	}

	// Create router and register commands with bot method bindings.
	router := NewRouter()
	router.RegisterCommand("start", bot.handleStartCmd)
	router.RegisterCommand("help", bot.handleHelpCmd)
	router.RegisterCommand("training", trainingH.HandleCommand)
	router.RegisterCommand("jointrun", jointrunH.HandleCommand)
	router.RegisterCallbacks(bot.handleCallback)
	bot.router = router

	return bot
}

// Start begins the polling loop.
func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.AllowedUpdates = []string{"message", "callback_query", "chat_member"}
	updates := b.api.GetUpdatesChan(u)

	b.logger.Info("telegram bot started polling")

	for {
		select {
		case <-b.stopCh:
			b.logger.Info("telegram bot stopped")
			return
		case update := <-updates:
			if err := b.processUpdate(update); err != nil {
				b.logger.Error("error processing update", zap.Error(err))
			}
		}
	}
}

// Stop signals the bot to stop polling.
func (b *Bot) Stop() {
	close(b.stopCh)
	b.api.StopReceivingUpdates()
	b.logger.Info("telegram bot stopping")
}

// ConfirmationHandler returns the confirmation handler for use by the scheduler.
func (b *Bot) ConfirmationHandler() *handlers.ConfirmationHandler {
	return b.confirmationH
}

// SendConfirmationPrompt sends a confirmation prompt to trainers (implements scheduler.BotMessenger).
func (b *Bot) SendConfirmationPrompt(ctx context.Context, trainingID int64) error {
	return b.confirmationH.SendConfirmationPrompt(ctx, trainingID)
}

// SendMessage sends a plain text message (implements usecase.MessageSender).
func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err := b.api.Send(msg)
	return err
}

// SendMessageWithKeyboard sends a text message with an inline keyboard (implements usecase.MessageSender).
func (b *Bot) SendMessageWithKeyboard(chatID int64, text string, keyboard any) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if kb, ok := keyboard.(tgbotapi.InlineKeyboardMarkup); ok {
		msg.ReplyMarkup = kb
	}
	_, err := b.api.Send(msg)
	return err
}

// handleStartCmd handles the /start command.
func (b *Bot) handleStartCmd(_ context.Context, msg *tgbotapi.Message) error {
	text := tgrescape.EscapeMarkdownV2("Hello! I'm the RunClub bot. Use /help to see available commands.")
	_ = b.SendMessage(msg.Chat.ID, text)
	return nil
}

// handleHelpCmd handles the /help command.
func (b *Bot) handleHelpCmd(_ context.Context, msg *tgbotapi.Message) error {
	text := tgrescape.EscapeMarkdownV2(
		"Available commands:\n/training - Create a new training\n/jointrun - Create a joint run",
	)
	_ = b.SendMessage(msg.Chat.ID, text)
	return nil
}

// processUpdate handles a single update from Telegram.
func (b *Bot) processUpdate(update tgbotapi.Update) error {
	ctx := context.Background()

	// Handle ChatMember updates (user joined group).
	if update.ChatMember != nil {
		return b.welcomeH.HandleChatMemberUpdate(ctx, update)
	}

	// Handle callback queries with special routing.
	if update.CallbackQuery != nil {
		return b.handleCallback(ctx, update.CallbackQuery)
	}

	// Handle photo messages in FSM context (e.g., training confirmation photo upload).
	if update.Message != nil && len(update.Message.Photo) > 0 {
		return b.stateMachine.Process(ctx, update.Message)
	}

	// Route commands and text messages.
	return b.router.Route(update, func(msg *tgbotapi.Message) error {
		// FSM handler for text messages.
		return b.stateMachine.Process(ctx, msg)
	})
}

// handleCallback dispatches callback queries to the appropriate handler based on callback data.
func (b *Bot) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	data := tgutil.Decode(cb.Data)
	if len(data) < 2 {
		return nil
	}

	action := data[0]
	entityType := data[1]

	switch {
	// Participation callbacks: join/leave training or joint run.
	case (action == handlers.ActionJoin || action == handlers.ActionLeave) &&
		(entityType == handlers.EntityTraining || entityType == handlers.EntityJointRun):
		return b.participationH.HandleCallback(ctx, cb)

	// Welcome callbacks.
	case action == handlers.ActionWelcome:
		return b.welcomeH.HandleWelcomeCallback(ctx, cb)

	// Race registration callbacks.
	case action == handlers.ActionRace && entityType == handlers.ActionRaceReg:
		return b.racePollH.HandleCallback(ctx, cb)

	// Confirmation callbacks: cfm/photo for training.
	case (action == handlers.ActionConfirm || action == handlers.ActionPhoto) &&
		entityType == handlers.EntityTraining:
		return b.confirmationH.HandleCallback(ctx, cb)

	// Add/remove participant during confirmation.
	case (action == handlers.ActionAddSelect || action == handlers.ActionRemSelect) &&
		entityType == handlers.EntityTraining && len(data) == 4:
		return b.handleConfirmationParticipant(ctx, cb, action, data)

	// Trainer selection / done callbacks (training creation).
	case (action == handlers.ActionAddSelect || action == handlers.ActionRemSelect) &&
		entityType == handlers.EntityTrainer:
		return b.trainingH.HandleCallback(ctx, cb)
	case action == handlers.ActionDone && entityType == handlers.EntityTrainer:
		return b.trainingH.HandleCallback(ctx, cb)

	// Club/location selection: disambiguate between training and joint run flows.
	case action == handlers.ActionSelect && entityType == handlers.EntityClub,
		action == handlers.ActionSelect && entityType == handlers.EntityLocation,
		action == handlers.ActionNew && entityType == handlers.EntityLocation:
		return b.handleClubLocationCallback(ctx, cb)

	default:
		return nil
	}
}

// handleConfirmationParticipant parses IDs and dispatches to add/remove participant callback.
func (b *Bot) handleConfirmationParticipant(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
	action string,
	data []string,
) error {
	memberID, err := strconv.ParseInt(data[3], 10, 64)
	if err != nil {
		return err
	}
	trainingID, err := strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		return err
	}
	if action == handlers.ActionAddSelect {
		return b.confirmationH.HandleAddParticipantCallback(ctx, cb, trainingID, memberID)
	}
	return b.confirmationH.HandleRemoveParticipantCallback(ctx, cb, trainingID, memberID)
}

// handleClubLocationCallback routes club/location selection to the active FSM handler.
func (b *Bot) handleClubLocationCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) error {
	if cb.Message != nil {
		state, err := b.stateMachine.GetState(ctx, cb.From.ID, cb.Message.Chat.ID, entity.FlowJointRunCreate)
		if err == nil && state != nil {
			return b.jointrunH.HandleCallback(ctx, cb)
		}
	}
	return b.trainingH.HandleCallback(ctx, cb)
}
