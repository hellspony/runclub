package entity

import "time"

type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "superadmin"
	AdminRoleAdmin      AdminRole = "admin"
)

type AdminUser struct {
	ID           int64     `json:"id"         db:"id"`
	Username     string    `json:"username"   db:"username"`
	PasswordHash string    `json:"-"          db:"password_hash"`
	Role         AdminRole `json:"role"       db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
