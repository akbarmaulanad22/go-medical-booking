package repository

import (
	"context"

	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DoctorProfileRepository interface {
	Create(ctx context.Context, db *gorm.DB, profile *entity.DoctorProfile) error
	FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (*entity.DoctorProfile, error)
}

type PatientProfileRepository interface {
	Create(ctx context.Context, db *gorm.DB, profile *entity.PatientProfile) error
	FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (*entity.PatientProfile, error)
}
