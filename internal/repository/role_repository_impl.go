package repository

import (
	"context"
	"errors"

	"go-template-clean-architecture/internal/domain/entity"
	domainRepo "go-template-clean-architecture/internal/domain/repository"

	"gorm.io/gorm"
)

type roleRepository struct{}

func NewRoleRepository() domainRepo.RoleRepository {
	return &roleRepository{}
}

func (r *roleRepository) FindByName(ctx context.Context, db *gorm.DB, name string) (*entity.Role, error) {
	var role entity.Role
	err := db.WithContext(ctx).Where("role_name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}
