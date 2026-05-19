package entity

import "time"

type Location struct {
	ID          int64     `json:"id"          db:"id"`
	ClubID      int64     `json:"club_id"     db:"club_id"`
	Name        string    `json:"name"        db:"name"`
	Address     string    `json:"address"     db:"address"`
	MapURL      string    `json:"map_url"     db:"map_url"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"  db:"updated_at"`
}
