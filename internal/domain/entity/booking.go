package entity

import (
	"time"

	"github.com/google/uuid"
)

// BookingStatus represents the status of a booking
type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// Booking represents a patient booking transaction
type Booking struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PatientID   uuid.UUID     `gorm:"type:uuid;not null;index" json:"patient_id"`
	ScheduleID  int           `gorm:"not null;index" json:"schedule_id"`
	BookingCode string        `gorm:"type:varchar(50);uniqueIndex;not null" json:"booking_code"`
	Status      BookingStatus `gorm:"type:booking_status;not null;default:'pending';index" json:"status"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Patient  PatientProfile `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Schedule DoctorSchedule `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
}

func (Booking) TableName() string {
	return "bookings"
}

// IsPending checks if booking is in pending status
func (b *Booking) IsPending() bool {
	return b.Status == BookingStatusPending
}

// IsConfirmed checks if booking is confirmed
func (b *Booking) IsConfirmed() bool {
	return b.Status == BookingStatusConfirmed
}

// IsCancelled checks if booking is cancelled
func (b *Booking) IsCancelled() bool {
	return b.Status == BookingStatusCancelled
}

// Confirm changes booking status to confirmed
func (b *Booking) Confirm() {
	b.Status = BookingStatusConfirmed
}

// Cancel changes booking status to cancelled
func (b *Booking) Cancel() {
	b.Status = BookingStatusCancelled
}
