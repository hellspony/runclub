package sqlite

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type AdminUserClubRepositoryImpl struct {
	db *sqlx.DB
}

func NewAdminUserClubRepository(db *sqlx.DB) *AdminUserClubRepositoryImpl {
	return &AdminUserClubRepositoryImpl{db: db}
}

func (r *AdminUserClubRepositoryImpl) Add(ctx context.Context, adminUserID, clubID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO admin_user_clubs (admin_user_id, club_id) VALUES (?, ?)`,
		adminUserID, clubID,
	)
	if err != nil {
		return fmt.Errorf("add admin user club: %w", err)
	}
	return nil
}

func (r *AdminUserClubRepositoryImpl) Remove(ctx context.Context, adminUserID, clubID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM admin_user_clubs WHERE admin_user_id = ? AND club_id = ?`,
		adminUserID, clubID,
	)
	if err != nil {
		return fmt.Errorf("remove admin user club: %w", err)
	}
	return nil
}

func (r *AdminUserClubRepositoryImpl) ListByAdminUser(ctx context.Context, adminUserID int64) ([]int64, error) {
	var ids []int64
	err := r.db.SelectContext(ctx, &ids,
		`SELECT club_id FROM admin_user_clubs WHERE admin_user_id = ?`, adminUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("list admin user clubs: %w", err)
	}
	return ids, nil
}

func (r *AdminUserClubRepositoryImpl) ListByClub(ctx context.Context, clubID int64) ([]int64, error) {
	var ids []int64
	err := r.db.SelectContext(ctx, &ids,
		`SELECT admin_user_id FROM admin_user_clubs WHERE club_id = ?`, clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("list club admin users: %w", err)
	}
	return ids, nil
}
