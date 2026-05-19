package sqlite

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"runclub/internal/domain/entity"
)

type CustomFieldRepositoryImpl struct {
	db *sqlx.DB
}

func NewCustomFieldRepository(db *sqlx.DB) *CustomFieldRepositoryImpl {
	return &CustomFieldRepositoryImpl{db: db}
}

type customFieldRow struct {
	ID        int64  `db:"id"`
	ClubID    int64  `db:"club_id"`
	Name      string `db:"name"`
	Required  bool   `db:"required"`
	SortOrder int    `db:"sort_order"`
}

type CustomFieldValueRepositoryImpl struct {
	db *sqlx.DB
}

func NewCustomFieldValueRepository(db *sqlx.DB) *CustomFieldValueRepositoryImpl {
	return &CustomFieldValueRepositoryImpl{db: db}
}

type customFieldValueRow struct {
	ID            int64  `db:"id"`
	MemberID      int64  `db:"member_id"`
	CustomFieldID int64  `db:"custom_field_id"`
	Value         string `db:"value"`
}

// CustomFieldRepositoryImpl methods

func (r *CustomFieldRepositoryImpl) Create(ctx context.Context, cf *entity.CustomField) (int64, error) {
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO custom_fields (club_id, name, required, sort_order)
		 VALUES (:club_id, :name, :required, :sort_order)`,
		map[string]any{
			"club_id":    cf.ClubID,
			"name":       cf.Name,
			"required":   cf.Required,
			"sort_order": cf.SortOrder,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert custom field: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *CustomFieldRepositoryImpl) ListByClub(ctx context.Context, clubID int64) ([]entity.CustomField, error) {
	var rows []customFieldRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, club_id, name, required, sort_order FROM custom_fields WHERE club_id = ? ORDER BY sort_order, name`,
		clubID,
	)
	if err != nil {
		return nil, fmt.Errorf("list custom fields by club: %w", err)
	}
	result := make([]entity.CustomField, len(rows))
	for i := range rows {
		result[i] = *customFieldRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *CustomFieldRepositoryImpl) Update(ctx context.Context, cf *entity.CustomField) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE custom_fields SET club_id = :club_id, name = :name, required = :required, sort_order = :sort_order
		 WHERE id = :id`,
		map[string]any{
			"id":         cf.ID,
			"club_id":    cf.ClubID,
			"name":       cf.Name,
			"required":   cf.Required,
			"sort_order": cf.SortOrder,
		},
	)
	if err != nil {
		return fmt.Errorf("update custom field: %w", err)
	}
	return nil
}

func (r *CustomFieldRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM custom_fields WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete custom field: %w", err)
	}
	return nil
}

func customFieldRowToEntity(row *customFieldRow) *entity.CustomField {
	return &entity.CustomField{
		ID:        row.ID,
		ClubID:    row.ClubID,
		Name:      row.Name,
		Required:  row.Required,
		SortOrder: row.SortOrder,
	}
}

// CustomFieldValueRepositoryImpl methods

func (r *CustomFieldValueRepositoryImpl) CreateOrUpdate(ctx context.Context, cfv *entity.CustomFieldValue) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT OR REPLACE INTO custom_field_values (member_id, custom_field_id, value)
		 VALUES (:member_id, :custom_field_id, :value)`,
		map[string]any{
			"member_id":       cfv.MemberID,
			"custom_field_id": cfv.CustomFieldID,
			"value":           cfv.Value,
		},
	)
	if err != nil {
		return fmt.Errorf("create or update custom field value: %w", err)
	}
	return nil
}

func (r *CustomFieldValueRepositoryImpl) ListByMember(
	ctx context.Context,
	memberID int64,
) ([]entity.CustomFieldValue, error) {
	var rows []customFieldValueRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, member_id, custom_field_id, value FROM custom_field_values WHERE member_id = ?`, memberID)
	if err != nil {
		return nil, fmt.Errorf("list custom field values by member: %w", err)
	}
	result := make([]entity.CustomFieldValue, len(rows))
	for i := range rows {
		result[i] = *customFieldValueRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *CustomFieldValueRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM custom_field_values WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete custom field value: %w", err)
	}
	return nil
}

func customFieldValueRowToEntity(row *customFieldValueRow) *entity.CustomFieldValue {
	return &entity.CustomFieldValue{
		ID:            row.ID,
		MemberID:      row.MemberID,
		CustomFieldID: row.CustomFieldID,
		Value:         row.Value,
	}
}
