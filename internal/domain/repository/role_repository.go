package repository

import (
	"context"

	"go-template-clean-architecture/internal/domain/entity"

	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByName(ctx context.Context, db *gorm.DB, name string) (*entity.Role, error)
}
