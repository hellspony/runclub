package repository

import (
	"context"

	"runclub/internal/domain/entity"
)

type AdminUserRepository interface {
	Create(ctx context.Context, user *entity.AdminUser) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.AdminUser, error)
	GetByUsername(ctx context.Context, username string) (*entity.AdminUser, error)
	List(ctx context.Context) ([]entity.AdminUser, error)
	Update(ctx context.Context, user *entity.AdminUser) error
	Delete(ctx context.Context, id int64) error
}
