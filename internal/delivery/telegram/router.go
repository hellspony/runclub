package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CommandFunc is a handler for a bot command.
type CommandFunc func(ctx context.Context, msg *tgbotapi.Message) error

// CallbackFunc is a handler for a callback query.
type CallbackFunc func(ctx context.Context, cb *tgbotapi.CallbackQuery) error

// Router dispatches commands and callbacks to registered handlers.
type Router struct {
	commands  map[string]CommandFunc
	callbacks CallbackFunc
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		commands: make(map[string]CommandFunc),
	}
}

// RegisterCommand registers a handler for a given command string.
func (r *Router) RegisterCommand(cmd string, fn CommandFunc) {
	r.commands[cmd] = fn
}

// RegisterCallbacks registers the callback handler.
func (r *Router) RegisterCallbacks(fn CallbackFunc) {
	r.callbacks = fn
}

// Route dispatches an update to the appropriate handler.
// Commands are routed by name, callbacks go to the callback handler,
// and text messages (without a command) are forwarded to the FSM handler.
func (r *Router) Route(update tgbotapi.Update, fsmHandler func(msg *tgbotapi.Message) error) error {
	ctx := context.Background()

	// Handle callback queries.
	if update.CallbackQuery != nil {
		if r.callbacks != nil {
			return r.callbacks(ctx, update.CallbackQuery)
		}
		return nil
	}

	// Handle text messages.
	if update.Message == nil {
		return nil
	}

	msg := update.Message

	// Check if it is a command.
	if msg.IsCommand() {
		cmd := msg.Command()
		if fn, ok := r.commands[cmd]; ok {
			return fn(ctx, msg)
		}
		// Unknown command - ignore.
		return nil
	}

	// Regular text message - delegate to FSM handler.
	if fsmHandler != nil {
		return fsmHandler(msg)
	}

	return nil
}
