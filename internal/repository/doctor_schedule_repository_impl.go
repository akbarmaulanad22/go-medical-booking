package repository

import (
	"context"
	"errors"

	"go-template-clean-architecture/internal/domain/entity"
	domainRepo "go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type doctorScheduleRepository struct{}

func NewDoctorScheduleRepository() domainRepo.DoctorScheduleRepository {
	return &doctorScheduleRepository{}
}

func (r *doctorScheduleRepository) Create(ctx context.Context, db *gorm.DB, schedule *entity.DoctorSchedule) error {
	return db.WithContext(ctx).Create(schedule).Error
}

func (r *doctorScheduleRepository) FindByID(ctx context.Context, db *gorm.DB, id int) (*entity.DoctorSchedule, error) {
	var schedule entity.DoctorSchedule
	err := db.WithContext(ctx).Preload("Doctor").Preload("Doctor.User").Where("id = ?", id).First(&schedule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &schedule, nil
}

func (r *doctorScheduleRepository) FindByDoctorID(ctx context.Context, db *gorm.DB, doctorID uuid.UUID) ([]entity.DoctorSchedule, error) {
	var schedules []entity.DoctorSchedule
	err := db.WithContext(ctx).Where("doctor_id = ?", doctorID).Order("schedule_date ASC, start_time ASC").Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *doctorScheduleRepository) FindAll(ctx context.Context, db *gorm.DB) ([]entity.DoctorSchedule, error) {
	var schedules []entity.DoctorSchedule
	err := db.WithContext(ctx).Preload("Doctor").Preload("Doctor.User").Order("schedule_date ASC, start_time ASC").Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *doctorScheduleRepository) Update(ctx context.Context, db *gorm.DB, schedule *entity.DoctorSchedule) error {
	return db.WithContext(ctx).Save(schedule).Error
}

func (r *doctorScheduleRepository) Delete(ctx context.Context, db *gorm.DB, id int) error {
	return db.WithContext(ctx).Where("id = ?", id).Delete(&entity.DoctorSchedule{}).Error
}
