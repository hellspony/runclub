package entity

import "time"

type BotFlowType string

const (
	FlowTrainingCreate  BotFlowType = "training_create"
	FlowJointRunCreate  BotFlowType = "jointrun_create"
	FlowWelcomeCollect  BotFlowType = "welcome_collect"
	FlowTrainingConfirm BotFlowType = "training_confirm"
)

type BotState struct {
	ID         int64       `json:"id"          db:"id"`
	TelegramID int64       `json:"telegram_id" db:"telegram_id"`
	ChatID     int64       `json:"chat_id"     db:"chat_id"`
	Flow       BotFlowType `json:"flow"        db:"flow"`
	Step       int         `json:"step"        db:"step"`
	Payload    string      `json:"payload"     db:"payload"`
	UpdatedAt  time.Time   `json:"updated_at"  db:"updated_at"`
}

type TrainingCreatePayload struct {
	ClubID     string  `json:"club_id"`
	LocationID string  `json:"location_id"`
	Date       string  `json:"date"`
	Duration   int     `json:"duration"`
	TrainerIDs []int64 `json:"trainer_ids"`
}

type JointRunCreatePayload struct {
	ClubID     string `json:"club_id"`
	LocationID string `json:"location_id"`
	Date       string `json:"date"`
}

type WelcomeCollectPayload struct {
	ClubID       string           `json:"club_id"`
	MemberID     string           `json:"member_id"`
	FIO          string           `json:"fio"`
	BirthDate    string           `json:"birth_date"`
	CustomFields map[int64]string `json:"custom_fields"`
	CurrentField int64            `json:"current_field"`
	Declined     bool             `json:"declined"`
}

type TrainingConfirmPayload struct {
	TrainingID  int64   `json:"training_id"`
	AddedIDs    []int64 `json:"added_ids"`
	RemovedIDs  []int64 `json:"removed_ids"`
	PhotoFileID string  `json:"photo_file_id"`
}
