package repository

import (
	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(db *gorm.DB, user *entity.User) error
	FindByEmail(db *gorm.DB, email string) (*entity.User, error)
	FindByID(db *gorm.DB, id uuid.UUID) (*entity.User, error)
}
