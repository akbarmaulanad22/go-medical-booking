package repository

import (
	"errors"
	"go-template-clean-architecture/internal/domain/entity"
	domainRepo "go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type doctorProfileRepository struct{}

func NewDoctorProfileRepository() domainRepo.DoctorProfileRepository {
	return &doctorProfileRepository{}
}

func (r *doctorProfileRepository) Create(db *gorm.DB, profile *entity.DoctorProfile) error {
	return db.Create(profile).Error
}

func (r *doctorProfileRepository) FindByUserID(db *gorm.DB, doctorID uuid.UUID) (*entity.DoctorProfile, error) {
	var profile entity.DoctorProfile
	err := db.Preload("User").Where("user_id = ?", doctorID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *doctorProfileRepository) FindAll(db *gorm.DB) ([]entity.DoctorProfile, error) {
	var profiles []entity.DoctorProfile
	err := db.Preload("User").Find(&profiles).Error
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (r *doctorProfileRepository) Update(db *gorm.DB, profile *entity.DoctorProfile) error {
	return db.Save(profile).Error
}

func (r *doctorProfileRepository) Delete(db *gorm.DB, doctorID uuid.UUID) error {
	return db.Where("user_id = ?", doctorID).Delete(&entity.DoctorProfile{}).Error
}
