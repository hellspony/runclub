package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"runclub/internal/domain/entity"
)

type AdminUserRepositoryImpl struct {
	db *sqlx.DB
}

func NewAdminUserRepository(db *sqlx.DB) *AdminUserRepositoryImpl {
	return &AdminUserRepositoryImpl{db: db}
}

type adminUserRow struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"`
	CreatedAt    string `db:"created_at"`
}

func (r *AdminUserRepositoryImpl) Create(ctx context.Context, user *entity.AdminUser) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO admin_users (username, password_hash, role, created_at)
		 VALUES (:username, :password_hash, :role, :created_at)`,
		map[string]any{
			"username":      user.Username,
			"password_hash": user.PasswordHash,
			"role":          string(user.Role),
			"created_at":    now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert admin user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *AdminUserRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.AdminUser, error) {
	var row adminUserRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, username, password_hash, role, created_at FROM admin_users WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("admin user not found: %w", err)
		}
		return nil, fmt.Errorf("get admin user by id: %w", err)
	}
	return adminUserRowToEntity(&row), nil
}

func (r *AdminUserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*entity.AdminUser, error) {
	var row adminUserRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, username, password_hash, role, created_at FROM admin_users WHERE username = ?`,
		username,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("admin user not found: %w", err)
		}
		return nil, fmt.Errorf("get admin user by username: %w", err)
	}
	return adminUserRowToEntity(&row), nil
}

func (r *AdminUserRepositoryImpl) List(ctx context.Context) ([]entity.AdminUser, error) {
	var rows []adminUserRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, username, password_hash, role, created_at FROM admin_users ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list admin users: %w", err)
	}
	result := make([]entity.AdminUser, len(rows))
	for i := range rows {
		result[i] = *adminUserRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *AdminUserRepositoryImpl) Update(ctx context.Context, user *entity.AdminUser) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE admin_users SET username = :username, role = :role WHERE id = :id`,
		map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"role":     string(user.Role),
		},
	)
	if err != nil {
		return fmt.Errorf("update admin user: %w", err)
	}
	return nil
}

func (r *AdminUserRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete admin user: %w", err)
	}
	return nil
}

func adminUserRowToEntity(row *adminUserRow) *entity.AdminUser {
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	return &entity.AdminUser{
		ID:           row.ID,
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Role:         entity.AdminRole(row.Role),
		CreatedAt:    createdAt,
	}
}
