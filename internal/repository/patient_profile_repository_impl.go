package repository

import (
	"context"
	"errors"

	"go-template-clean-architecture/internal/domain/entity"
	domainRepo "go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type patientProfileRepository struct{}

func NewPatientProfileRepository() domainRepo.PatientProfileRepository {
	return &patientProfileRepository{}
}

func (r *patientProfileRepository) Create(ctx context.Context, db *gorm.DB, profile *entity.PatientProfile) error {
	return db.WithContext(ctx).Create(profile).Error
}

func (r *patientProfileRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (*entity.PatientProfile, error) {
	var profile entity.PatientProfile
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *patientProfileRepository) FindAll(ctx context.Context, db *gorm.DB) ([]entity.PatientProfile, error) {
	var profiles []entity.PatientProfile
	err := db.WithContext(ctx).Preload("User").Find(&profiles).Error
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (r *patientProfileRepository) Update(ctx context.Context, db *gorm.DB, profile *entity.PatientProfile) error {
	return db.WithContext(ctx).Save(profile).Error
}

func (r *patientProfileRepository) Delete(ctx context.Context, db *gorm.DB, userID uuid.UUID) error {
	return db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.PatientProfile{}).Error
}
