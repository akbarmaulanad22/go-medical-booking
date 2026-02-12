package repository

import (
	"errors"

	"go-template-clean-architecture/internal/domain/entity"
	domainRepo "go-template-clean-architecture/internal/domain/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bookingRepository struct{}

func NewBookingRepository() domainRepo.BookingRepository {
	return &bookingRepository{}
}

func (r *bookingRepository) Create(db *gorm.DB, booking *entity.Booking) error {
	return db.Create(booking).Error
}

func (r *bookingRepository) FindByID(db *gorm.DB, id uuid.UUID) (*entity.Booking, error) {
	var booking entity.Booking
	err := db.Preload("Schedule.Doctor").Where("id = ?", id).First(&booking).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) FindByPatientID(db *gorm.DB, patientID uuid.UUID) ([]entity.Booking, error) {
	var bookings []entity.Booking
	err := db.Preload("Schedule.Doctor").
		Where("patient_id = ?", patientID).
		Order("created_at DESC").
		Find(&bookings).Error
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

// CancelBooking atomically cancels a booking ONLY if it's not already cancelled.
// Returns affected rows: 1 = success, 0 = already cancelled (prevents double-cancel race).
func (r *bookingRepository) CancelBooking(db *gorm.DB, id uuid.UUID) (int64, error) {
	result := db.Model(&entity.Booking{}).
		Where("id = ? AND status != ?", id, entity.BookingStatusCancelled).
		Update("status", entity.BookingStatusCancelled)
	return result.RowsAffected, result.Error
}

func (r *bookingRepository) FindByPatientAndSchedule(db *gorm.DB, patientID uuid.UUID, scheduleID int) (*entity.Booking, error) {
	var booking entity.Booking
	err := db.Where("patient_id = ? AND schedule_id = ? AND status != ?", patientID, scheduleID, entity.BookingStatusCancelled).
		First(&booking).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}
