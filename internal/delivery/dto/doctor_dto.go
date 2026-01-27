package dto

import (
	"github.com/google/uuid"
)

// Request DTOs

type CreateDoctorRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	FullName       string `json:"full_name" validate:"required,min=2"`
	STRNumber      string `json:"str_number" validate:"required"`
	Specialization string `json:"specialization" validate:"required"`
	Biography      string `json:"biography" validate:"omitempty"`
}

type UpdateDoctorRequest struct {
	Email          string `json:"email" validate:"omitempty,email"`
	Password       string `json:"password" validate:"omitempty,min=6"`
	FullName       string `json:"full_name" validate:"omitempty,min=2"`
	STRNumber      string `json:"str_number" validate:"omitempty"`
	Specialization string `json:"specialization" validate:"omitempty"`
	Biography      string `json:"biography" validate:"omitempty"`
	IsActive       *bool  `json:"is_active" validate:"omitempty"`
}

// Response DTOs

type DoctorResponse struct {
	ID             uuid.UUID `json:"id"`
	Email          string    `json:"email"`
	FullName       string    `json:"full_name"`
	STRNumber      string    `json:"str_number"`
	Specialization string    `json:"specialization"`
	Biography      string    `json:"biography,omitempty"`
	IsActive       *bool     `json:"is_active"`
}

type DoctorListResponse struct {
	Doctors []DoctorResponse `json:"doctors"`
	Total   int              `json:"total"`
}
