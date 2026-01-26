package entity

import (
	"time"

	"github.com/google/uuid"
)

// PatientProfile represents patient-specific profile data
type PatientProfile struct {
	UserID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	NIK         string    `gorm:"type:char(16);uniqueIndex;not null" json:"nik"`
	PhoneNumber string    `gorm:"type:varchar(20);index" json:"phone_number,omitempty"`
	DateOfBirth time.Time `gorm:"type:date;not null" json:"date_of_birth"`
	Gender      string    `gorm:"type:char(1);not null" json:"gender"`
	Address     string    `gorm:"type:text" json:"address,omitempty"`

	// Relationships
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Bookings []Booking `gorm:"foreignKey:PatientID" json:"bookings,omitempty"`
}

func (PatientProfile) TableName() string {
	return "patient_profiles"
}

// Gender constants
const (
	GenderMale   = "M"
	GenderFemale = "F"
)
