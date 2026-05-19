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

type LocationRepositoryImpl struct {
	db *sqlx.DB
}

func NewLocationRepository(db *sqlx.DB) *LocationRepositoryImpl {
	return &LocationRepositoryImpl{db: db}
}

type locationRow struct {
	ID          int64  `db:"id"`
	ClubID      int64  `db:"club_id"`
	Name        string `db:"name"`
	Address     string `db:"address"`
	MapURL      string `db:"map_url"`
	Description string `db:"description"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}

func (r *LocationRepositoryImpl) Create(ctx context.Context, location *entity.Location) (int64, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO locations (club_id, name, address, map_url, description, created_at, updated_at)
		 VALUES (:club_id, :name, :address, :map_url, :description, :created_at, :updated_at)`,
		map[string]any{
			"club_id":     location.ClubID,
			"name":        location.Name,
			"address":     location.Address,
			"map_url":     location.MapURL,
			"description": location.Description,
			"created_at":  now,
			"updated_at":  now,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert location: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *LocationRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Location, error) {
	var row locationRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, club_id, name, address, map_url, description, created_at, updated_at FROM locations WHERE id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("location not found: %w", err)
		}
		return nil, fmt.Errorf("get location by id: %w", err)
	}
	return locationRowToEntity(&row), nil
}

func (r *LocationRepositoryImpl) ListByClub(ctx context.Context, clubID int64) ([]entity.Location, error) {
	var rows []locationRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, club_id, name, address, map_url, description, created_at, updated_at FROM locations WHERE club_id = ? ORDER BY name`,
		clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("list locations by club: %w", err)
	}
	result := make([]entity.Location, len(rows))
	for i := range rows {
		result[i] = *locationRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *LocationRepositoryImpl) Update(ctx context.Context, location *entity.Location) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE locations SET club_id = :club_id, name = :name, address = :address,
		 map_url = :map_url, description = :description, updated_at = :updated_at
		 WHERE id = :id`,
		map[string]any{
			"id":          location.ID,
			"club_id":     location.ClubID,
			"name":        location.Name,
			"address":     location.Address,
			"map_url":     location.MapURL,
			"description": location.Description,
			"updated_at":  now,
		},
	)
	if err != nil {
		return fmt.Errorf("update location: %w", err)
	}
	return nil
}

func (r *LocationRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM locations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete location: %w", err)
	}
	return nil
}

func locationRowToEntity(row *locationRow) *entity.Location {
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, row.UpdatedAt)
	return &entity.Location{
		ID:          row.ID,
		ClubID:      row.ClubID,
		Name:        row.Name,
		Address:     row.Address,
		MapURL:      row.MapURL,
		Description: row.Description,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
