package entity

import "time"

type TrainingStatus string

const (
	TrainingStatusPlanned    TrainingStatus = "planned"
	TrainingStatusInProgress TrainingStatus = "in_progress"
	TrainingStatusConfirming TrainingStatus = "confirming"
	TrainingStatusCompleted  TrainingStatus = "completed"
)

type Training struct {
	ID          int64          `json:"id"            db:"id"`
	ClubID      int64          `json:"club_id"       db:"club_id"`
	LocationID  int64          `json:"location_id"   db:"location_id"`
	Date        time.Time      `json:"date"          db:"date"`
	Duration    int            `json:"duration"      db:"duration"`
	Status      TrainingStatus `json:"status"        db:"status"`
	PhotoFileID string         `json:"photo_file_id" db:"photo_file_id"`
	MessageID   int64          `json:"message_id"    db:"message_id"`
	CreatedAt   time.Time      `json:"created_at"    db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"    db:"updated_at"`
}

type TrainingTrainer struct {
	ID         int64 `json:"id"          db:"id"`
	TrainingID int64 `json:"training_id" db:"training_id"`
	MemberID   int64 `json:"member_id"   db:"member_id"`
}

type TrainingParticipant struct {
	ID         int64 `json:"id"          db:"id"`
	TrainingID int64 `json:"training_id" db:"training_id"`
	MemberID   int64 `json:"member_id"   db:"member_id"`
}
