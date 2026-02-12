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

// FindAllWithActiveDoctor returns schedules only for doctors whose user account is active.
// Supports optional filters: date range, doctor name, and specialization.
func (r *doctorScheduleRepository) FindAllWithActiveDoctor(db *gorm.DB, filter *entity.ScheduleFilter) ([]entity.DoctorSchedule, error) {
	var schedules []entity.DoctorSchedule
	query := db.
		Joins("JOIN doctor_profiles ON doctor_profiles.user_id = doctor_schedules.doctor_id").
		Joins("JOIN users ON users.id = doctor_profiles.user_id").
		Where("users.is_active = ?", true)

	if filter != nil {
		if filter.StartAt != "" {
			query = query.Where("doctor_schedules.schedule_date >= ?", filter.StartAt)
		}
		if filter.EndAt != "" {
			query = query.Where("doctor_schedules.schedule_date <= ?", filter.EndAt)
		}
		if filter.DoctorName != "" {
			query = query.Where("users.full_name ILIKE ?", "%"+filter.DoctorName+"%")
		}
		if filter.Specialization != "" {
			query = query.Where("doctor_profiles.specialization ILIKE ?", "%"+filter.Specialization+"%")
		}
	}

	err := query.
		Preload("Doctor").Preload("Doctor.User").
		Order("schedule_date ASC, start_time ASC").
		Find(&schedules).Error
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
