package dto

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs

type CreateBookingRequest struct {
	ScheduleID int `json:"schedule_id" validate:"required,min=1"`
}

// Response DTOs

type BookingResponse struct {
	ID          uuid.UUID         `json:"id"`
	PatientID   uuid.UUID         `json:"patient_id"`
	ScheduleID  int               `json:"schedule_id"`
	BookingCode string            `json:"booking_code"`
	QueueNumber int               `json:"queue_number"`
	Status      string            `json:"status"`
	Schedule    *ScheduleResponse `json:"schedule,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type BookingListResponse struct {
	Bookings []BookingResponse `json:"bookings"`
	Total    int               `json:"total"`
}
