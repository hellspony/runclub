package repository

import (
	"context"

	"runclub/internal/domain/entity"
)

type TemplateRepository interface {
	Create(ctx context.Context, tmpl *entity.Template) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Template, error)
	GetByClubAndType(ctx context.Context, clubID int64, tmplType entity.TemplateType) (*entity.Template, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.Template, error)
	Update(ctx context.Context, tmpl *entity.Template) error
	Delete(ctx context.Context, id int64) error
}
