package repository

import (
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

func (r *doctorScheduleRepository) Create(db *gorm.DB, schedule *entity.DoctorSchedule) error {
	return db.Create(schedule).Error
}

func (r *doctorScheduleRepository) FindByID(db *gorm.DB, id int) (*entity.DoctorSchedule, error) {
	var schedule entity.DoctorSchedule
	err := db.Preload("Doctor.User").Where("id = ?", id).First(&schedule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &schedule, nil
}

func (r *doctorScheduleRepository) FindByDoctorID(db *gorm.DB, doctorID uuid.UUID) ([]entity.DoctorSchedule, error) {
	var schedules []entity.DoctorSchedule
	err := db.Where("doctor_id = ?", doctorID).Order("schedule_date ASC, start_time ASC").Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *doctorScheduleRepository) FindAll(db *gorm.DB) ([]entity.DoctorSchedule, error) {
	var schedules []entity.DoctorSchedule
	err := db.Preload("Doctor").Preload("Doctor.User").Order("schedule_date ASC, start_time ASC").Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *doctorScheduleRepository) Update(db *gorm.DB, schedule *entity.DoctorSchedule) error {
	return db.Omit("Doctor").Save(schedule).Error
}

func (r *doctorScheduleRepository) Delete(db *gorm.DB, id int) (int64, error) {
	affected := db.Where("id = ?", id).Delete(&entity.DoctorSchedule{})
	return affected.RowsAffected, affected.Error

}
