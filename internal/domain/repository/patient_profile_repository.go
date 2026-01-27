package repository

import (
	"context"

	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PatientProfileRepository interface {
	Create(ctx context.Context, db *gorm.DB, profile *entity.PatientProfile) error
	FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (*entity.PatientProfile, error)
	FindAll(ctx context.Context, db *gorm.DB) ([]entity.PatientProfile, error)
	Update(ctx context.Context, db *gorm.DB, profile *entity.PatientProfile) error
	Delete(ctx context.Context, db *gorm.DB, userID uuid.UUID) error
}
