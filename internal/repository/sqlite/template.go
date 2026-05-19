package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"runclub/internal/domain/entity"
)

type TemplateRepositoryImpl struct {
	db *sqlx.DB
}

func NewTemplateRepository(db *sqlx.DB) *TemplateRepositoryImpl {
	return &TemplateRepositoryImpl{db: db}
}

type templateRow struct {
	ID      int64  `db:"id"`
	ClubID  int64  `db:"club_id"`
	Type    string `db:"type"`
	Name    string `db:"name"`
	Content string `db:"content"`
}

func (r *TemplateRepositoryImpl) Create(ctx context.Context, tmpl *entity.Template) (int64, error) {
	result, err := r.db.NamedExecContext(ctx,
		`INSERT INTO templates (club_id, type, name, content)
		 VALUES (:club_id, :type, :name, :content)`,
		map[string]any{
			"club_id": tmpl.ClubID,
			"type":    tmpl.Type,
			"name":    tmpl.Name,
			"content": tmpl.Content,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("insert template: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

func (r *TemplateRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.Template, error) {
	var row templateRow
	err := r.db.GetContext(ctx, &row, `SELECT id, club_id, type, name, content FROM templates WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("template not found: %w", err)
		}
		return nil, fmt.Errorf("get template by id: %w", err)
	}
	return templateRowToEntity(&row), nil
}

func (r *TemplateRepositoryImpl) GetByClubAndType(
	ctx context.Context,
	clubID int64,
	tmplType entity.TemplateType,
) (*entity.Template, error) {
	var row templateRow
	err := r.db.GetContext(ctx, &row,
		`SELECT id, club_id, type, name, content FROM templates WHERE club_id = ? AND type = ?`, clubID, tmplType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("template not found: %w", err)
		}
		return nil, fmt.Errorf("get template by club and type: %w", err)
	}
	return templateRowToEntity(&row), nil
}

func (r *TemplateRepositoryImpl) ListByClub(ctx context.Context, clubID int64) ([]entity.Template, error) {
	var rows []templateRow
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, club_id, type, name, content FROM templates WHERE club_id = ? ORDER BY type`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list templates by club: %w", err)
	}
	result := make([]entity.Template, len(rows))
	for i := range rows {
		result[i] = *templateRowToEntity(&rows[i])
	}
	return result, nil
}

func (r *TemplateRepositoryImpl) Update(ctx context.Context, tmpl *entity.Template) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE templates SET club_id = :club_id, type = :type, name = :name, content = :content
		 WHERE id = :id`,
		map[string]any{
			"id":      tmpl.ID,
			"club_id": tmpl.ClubID,
			"type":    tmpl.Type,
			"name":    tmpl.Name,
			"content": tmpl.Content,
		},
	)
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}
	return nil
}

func (r *TemplateRepositoryImpl) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM templates WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

func templateRowToEntity(row *templateRow) *entity.Template {
	return &entity.Template{
		ID:      row.ID,
		ClubID:  row.ClubID,
		Type:    entity.TemplateType(row.Type),
		Name:    row.Name,
		Content: row.Content,
	}
}
