package dto

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Response DTOs

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserResponse struct {
	ID             uuid.UUID               `json:"id"`
	Email          string                  `json:"email"`
	FullName       string                  `json:"full_name"`
	Role           string                  `json:"role"`
	DoctorProfile  *DoctorProfileResponse  `json:"doctor_profile,omitempty"`
	PatientProfile *PatientProfileResponse `json:"patient_profile,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

// Role-specific Registration Request DTOs

// RegisterPatientRequest untuk registrasi pasien
type RegisterPatientRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	FullName    string `json:"full_name" validate:"required,min=2"`
	NIK         string `json:"nik" validate:"required,len=16"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,min=10,max=20"`
	DateOfBirth string `json:"date_of_birth" validate:"required"` // Format: YYYY-MM-DD
	Gender      string `json:"gender" validate:"required,oneof=M F"`
	Address     string `json:"address" validate:"omitempty"`
}

// RegisterDoctorRequest untuk registrasi dokter
type RegisterDoctorRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	FullName       string `json:"full_name" validate:"required,min=2"`
	STRNumber      string `json:"str_number" validate:"required"`
	Specialization string `json:"specialization" validate:"required"`
	Biography      string `json:"biography" validate:"omitempty"`
}
