package repository

import (
	"context"

	"runclub/internal/domain/entity"
)

type CustomFieldRepository interface {
	Create(ctx context.Context, cf *entity.CustomField) (int64, error)
	ListByClub(ctx context.Context, clubID int64) ([]entity.CustomField, error)
	Update(ctx context.Context, cf *entity.CustomField) error
	Delete(ctx context.Context, id int64) error
}

type CustomFieldValueRepository interface {
	CreateOrUpdate(ctx context.Context, cfv *entity.CustomFieldValue) error
	ListByMember(ctx context.Context, memberID int64) ([]entity.CustomFieldValue, error)
	Delete(ctx context.Context, id int64) error
}
