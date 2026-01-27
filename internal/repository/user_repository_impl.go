package repository

import (
	"go-template-clean-architecture/internal/domain/entity"
	domainRepo "go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct{}

func NewUserRepository() domainRepo.UserRepository {
	return &userRepository{}
}

func (r *userRepository) Create(db *gorm.DB, user *entity.User) error {
	return db.Create(user).Error
}

func (r *userRepository) FindByEmail(db *gorm.DB, email string) (*entity.User, error) {
	var user entity.User
	err := db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID(db *gorm.DB, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := db.Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) Update(db *gorm.DB, user *entity.User) error {
	return db.Save(user).Error
}
