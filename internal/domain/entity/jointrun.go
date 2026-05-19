package entity

import "time"

type JointRun struct {
	ID         int64     `json:"id"          db:"id"`
	ClubID     int64     `json:"club_id"     db:"club_id"`
	LocationID int64     `json:"location_id" db:"location_id"`
	CreatorID  int64     `json:"creator_id"  db:"creator_id"`
	Date       time.Time `json:"date"        db:"date"`
	MessageID  int64     `json:"message_id"  db:"message_id"`
	CreatedAt  time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"  db:"updated_at"`
}

type JointRunParticipant struct {
	ID         int64 `json:"id"           db:"id"`
	JointRunID int64 `json:"joint_run_id" db:"joint_run_id"`
	MemberID   int64 `json:"member_id"    db:"member_id"`
}
