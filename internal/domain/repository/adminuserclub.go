package repository

import "context"

type AdminUserClubRepository interface {
	Add(ctx context.Context, adminUserID, clubID int64) error
	Remove(ctx context.Context, adminUserID, clubID int64) error
	ListByAdminUser(ctx context.Context, adminUserID int64) ([]int64, error)
	ListByClub(ctx context.Context, clubID int64) ([]int64, error)
}
