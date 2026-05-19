package entity

import "time"

type Race struct {
	ID        int64     `json:"id"         db:"id"`
	ClubID    int64     `json:"club_id"    db:"club_id"`
	Date      time.Time `json:"date"       db:"date"`
	Type      string    `json:"type"       db:"type"`
	Place     string    `json:"place"      db:"place"`
	Distances string    `json:"distances"  db:"distances"`
	Name      string    `json:"name"       db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RaceRegistration struct {
	ID        int64     `json:"id"         db:"id"`
	RaceID    int64     `json:"race_id"    db:"race_id"`
	MemberID  int64     `json:"member_id"  db:"member_id"`
	Distance  string    `json:"distance"   db:"distance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
