package entity

import "time"

type Club struct {
	ID                int64     `json:"id"                  db:"id"`
	Name              string    `json:"name"                db:"name"`
	TelegramChatID    int64     `json:"telegram_chat_id"    db:"telegram_chat_id"`
	WelcomeEnabled    bool      `json:"welcome_enabled"     db:"welcome_enabled"`
	BirthdayEnabled   bool      `json:"birthday_enabled"    db:"birthday_enabled"`
	RaceNotifyEnabled bool      `json:"race_notify_enabled" db:"race_notify_enabled"`
	CreatedAt         time.Time `json:"created_at"          db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"          db:"updated_at"`
}
