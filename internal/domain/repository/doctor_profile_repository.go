package repository

import (
	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DoctorProfileRepository interface {
	Create(db *gorm.DB, profile *entity.DoctorProfile) error
	FindByUserID(db *gorm.DB, userID uuid.UUID) (*entity.DoctorProfile, error)
	FindAll(db *gorm.DB) ([]entity.DoctorProfile, error)
	Update(db *gorm.DB, profile *entity.DoctorProfile) error
	Delete(db *gorm.DB, userID uuid.UUID) error
}
