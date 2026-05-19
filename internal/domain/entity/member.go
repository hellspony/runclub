package entity

import "time"

type MemberRole string

const (
	RoleMember  MemberRole = "member"
	RoleTrainer MemberRole = "trainer"
	RoleAdmin   MemberRole = "admin"
)

type Member struct {
	ID               int64      `json:"id"                db:"id"`
	FIO              string     `json:"fio"               db:"fio"`
	TelegramUsername string     `json:"telegram_username" db:"telegram_username"`
	TelegramID       int64      `json:"telegram_id"       db:"telegram_id"`
	BirthDate        *time.Time `json:"birth_date"        db:"birth_date"`
	LeftAt           *time.Time `json:"left_at"           db:"left_at"`
	CreatedAt        time.Time  `json:"created_at"        db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"        db:"updated_at"`
}

type ClubMember struct {
	ID       int64      `json:"id"        db:"id"`
	ClubID   int64      `json:"club_id"   db:"club_id"`
	MemberID int64      `json:"member_id" db:"member_id"`
	Role     MemberRole `json:"role"      db:"role"`
	JoinedAt time.Time  `json:"joined_at" db:"joined_at"`
}
