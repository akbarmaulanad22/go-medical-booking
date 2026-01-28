package dto

import (
	"time"

	"github.com/google/uuid"
)

// PatientProfileResponse represents patient profile data in responses
type PatientProfileResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	NIK         string    `json:"nik"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	DateOfBirth string    `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Address     string    `json:"address,omitempty"`
}

// PatientResponse represents a patient user with profile data
type PatientResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	NIK         string    `json:"nik"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	DateOfBirth string    `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Address     string    `json:"address,omitempty"`
	IsActive    *bool     `json:"is_active,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
