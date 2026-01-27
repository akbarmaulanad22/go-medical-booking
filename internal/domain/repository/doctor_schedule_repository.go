package repository

import (
	"context"

	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DoctorScheduleRepository interface {
	Create(ctx context.Context, db *gorm.DB, schedule *entity.DoctorSchedule) error
	FindByID(ctx context.Context, db *gorm.DB, id int) (*entity.DoctorSchedule, error)
	FindByDoctorID(ctx context.Context, db *gorm.DB, doctorID uuid.UUID) ([]entity.DoctorSchedule, error)
	FindAll(ctx context.Context, db *gorm.DB) ([]entity.DoctorSchedule, error)
	Update(ctx context.Context, db *gorm.DB, schedule *entity.DoctorSchedule) error
	Delete(ctx context.Context, db *gorm.DB, id int) error
}
