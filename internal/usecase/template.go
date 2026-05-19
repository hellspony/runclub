package usecase

import (
	"context"

	"runclub/internal/domain/entity"
	"runclub/internal/domain/repository"
)

type TemplateUseCase interface {
	Create(ctx context.Context, tmpl *entity.Template) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Template, error)
	GetByClubAndType(ctx context.Context, clubID int64, tmplType entity.TemplateType) (*entity.Template, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.Template, error)
	Update(ctx context.Context, tmpl *entity.Template) error
	Delete(ctx context.Context, id int64) error
}

type templateUseCase struct {
	tmplRepo repository.TemplateRepository
}

func NewTemplateUseCase(tmplRepo repository.TemplateRepository) TemplateUseCase {
	return &templateUseCase{
		tmplRepo: tmplRepo,
	}
}

func (uc *templateUseCase) Create(ctx context.Context, tmpl *entity.Template) (int64, error) {
	return uc.tmplRepo.Create(ctx, tmpl)
}

func (uc *templateUseCase) GetByID(ctx context.Context, id int64) (*entity.Template, error) {
	return uc.tmplRepo.GetByID(ctx, id)
}

func (uc *templateUseCase) GetByClubAndType(
	ctx context.Context,
	clubID int64,
	tmplType entity.TemplateType,
) (*entity.Template, error) {
	return uc.tmplRepo.GetByClubAndType(ctx, clubID, tmplType)
}

func (uc *templateUseCase) ListByClub(ctx context.Context, clubID int64) ([]entity.Template, error) {
	return uc.tmplRepo.ListByClub(ctx, clubID)
}

func (uc *templateUseCase) Update(ctx context.Context, tmpl *entity.Template) error {
	return uc.tmplRepo.Update(ctx, tmpl)
}

func (uc *templateUseCase) Delete(ctx context.Context, id int64) error {
	return uc.tmplRepo.Delete(ctx, id)
}
