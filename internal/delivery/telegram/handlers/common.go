package handlers

import (
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Callback action constants used in inline keyboard data.
const (
	ActionSelect    = "sel"
	ActionNew       = "new"
	ActionAddSelect = "addsel"
	ActionRemSelect = "remsel"
	ActionJoin      = "join"
	ActionLeave     = "leave"
	ActionConfirm   = "cfm"
	ActionPhoto     = "photo"
	ActionDone      = "done"
	ActionWelcome   = "wel"
	ActionRace      = "race"
	ActionRaceReg   = "reg"

	EntityTraining = "training"
	EntityJointRun = "jointrun"
	EntityClub     = "club"
	EntityLocation = "loc"
	EntityTrainer  = "trainer"

	// TplLocationName and TplDate are template field names used in templater.Render calls.
	TplLocationName = "LocationName"
	TplDate         = "Date"
)

// sendText sends a plain text message to a chat.
func sendText(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err := bot.Send(msg)
	return err
}

// sendInlineKeyboard sends a text message with an inline keyboard.
func sendInlineKeyboard(bot *tgbotapi.BotAPI, chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	return err
}

// editMessageText edits an existing message text, optionally updating the inline keyboard.
func editMessageText(
	bot *tgbotapi.BotAPI,
	chatID int64,
	messageID int,
	text string,
	keyboard *tgbotapi.InlineKeyboardMarkup,
) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdownV2
	if keyboard != nil {
		edit.ReplyMarkup = keyboard
	}
	_, err := bot.Send(edit)
	return err
}

// answerCallback answers a callback query with optional text.
func answerCallback(bot *tgbotapi.BotAPI, callbackID, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)
	_, err := bot.Request(callback)
	return err
}

// parseDateTime tries multiple date/time formats and returns the parsed time.
func parseDateTime(text string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04",
		"02.01.2006 15:04",
		"2006-01-02T15:04",
		"02.01.2006T15:04",
		"2006-01-02 15:04:05",
		"02.01.2006 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, text); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date/time: %q", text)
}

// toggleID adds or removes an ID from a slice, returning the new slice.
func toggleID(ids []int64, id int64) []int64 {
	for i, existing := range ids {
		if existing == id {
			// Remove it.
			return append(ids[:i], ids[i+1:]...)
		}
	}
	// Add it.
	return append(ids, id)
}

// formatInt64 converts an int64 to string.
func formatInt64(v int64) string {
	return strconv.FormatInt(v, 10)
}

// parseInt64 parses a string as int64.
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
