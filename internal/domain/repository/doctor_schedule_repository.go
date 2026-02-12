package repository

import (
	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DoctorScheduleRepository interface {
	Create(db *gorm.DB, schedule *entity.DoctorSchedule) error
	FindByID(db *gorm.DB, id int) (*entity.DoctorSchedule, error)
	FindByDoctorID(db *gorm.DB, doctorID uuid.UUID) ([]entity.DoctorSchedule, error)
	FindAll(db *gorm.DB) ([]entity.DoctorSchedule, error)
	FindAllWithActiveDoctor(db *gorm.DB, filter *entity.ScheduleFilter) ([]entity.DoctorSchedule, error)
	Update(db *gorm.DB, schedule *entity.DoctorSchedule) error
	Delete(db *gorm.DB, id int) (int64, error)
}
