package repository

import (
	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(db *gorm.DB, booking *entity.Booking) error
	FindByID(db *gorm.DB, id uuid.UUID) (*entity.Booking, error)
	FindByPatientID(db *gorm.DB, patientID uuid.UUID) ([]entity.Booking, error)
	UpdateStatus(db *gorm.DB, id uuid.UUID, status entity.BookingStatus) error
	FindByPatientAndSchedule(db *gorm.DB, patientID uuid.UUID, scheduleID int) (*entity.Booking, error)
}
