package repository

import (
	"context"
	"errors"

	"go-template-clean-architecture/internal/domain/entity"
	domainRepo "go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Doctor Profile Repository

type doctorProfileRepository struct{}

func NewDoctorProfileRepository() domainRepo.DoctorProfileRepository {
	return &doctorProfileRepository{}
}

func (r *doctorProfileRepository) Create(ctx context.Context, db *gorm.DB, profile *entity.DoctorProfile) error {
	return db.WithContext(ctx).Create(profile).Error
}

func (r *doctorProfileRepository) FindByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) (*entity.DoctorProfile, error) {
	var profile entity.DoctorProfile
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// Patient Profile Repository

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
