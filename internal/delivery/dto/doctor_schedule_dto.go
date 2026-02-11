package dto

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs

type CreateScheduleRequest struct {
	DoctorID     uuid.UUID `json:"doctor_id" validate:"required"`
	ScheduleDate string    `json:"schedule_date" validate:"required"` // Format: YYYY-MM-DD
	StartTime    string    `json:"start_time" validate:"required"`    // Format: HH:MM
	EndTime      string    `json:"end_time" validate:"required"`      // Format: HH:MM
	TotalQuota   int       `json:"total_quota" validate:"required,min=1"`
}

type UpdateScheduleRequest struct {
	DoctorID     uuid.UUID `json:"doctor_id" validate:"omitempty"`
	ScheduleDate string    `json:"schedule_date" validate:"omitempty"` // Format: YYYY-MM-DD
	StartTime    string    `json:"start_time" validate:"omitempty"`    // Format: HH:MM
	EndTime      string    `json:"end_time" validate:"omitempty"`      // Format: HH:MM
	TotalQuota   *int      `json:"total_quota" validate:"omitempty,min=1"`
}

// Response DTOs

type ScheduleResponse struct {
	ID           int             `json:"id"`
	DoctorID     uuid.UUID       `json:"doctor_id"`
	Doctor       *DoctorResponse `json:"doctor,omitempty"`
	ScheduleDate string          `json:"schedule_date"`
	StartTime    string          `json:"start_time"`
	EndTime      string          `json:"end_time"`
	TotalQuota   int             `json:"total_quota"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type ScheduleListResponse struct {
	Schedules []ScheduleResponse `json:"schedules"`
	Total     int                `json:"total"`
}
