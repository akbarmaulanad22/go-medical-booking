package converter

import (
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
)

// BookingToResponse converts a Booking entity to BookingResponse DTO
func BookingToResponse(booking *entity.Booking) *dto.BookingResponse {
	if booking == nil {
		return nil
	}

	response := &dto.BookingResponse{
		ID:          booking.ID,
		PatientID:   booking.PatientID,
		ScheduleID:  booking.ScheduleID,
		BookingCode: booking.BookingCode,
		QueueNumber: booking.QueueNumber,
		Status:      string(booking.Status),
		CreatedAt:   booking.CreatedAt,
		UpdatedAt:   booking.UpdatedAt,
	}

	// Include schedule info if available
	if booking.Schedule.ID != 0 {
		response.Schedule = ScheduleToResponse(&booking.Schedule)
	}

	return response
}

// BookingsToResponses converts a slice of Booking entities to slice of BookingResponse DTOs
func BookingsToResponses(bookings []entity.Booking) []dto.BookingResponse {
	responses := make([]dto.BookingResponse, len(bookings))
	for i, booking := range bookings {
		resp := BookingToResponse(&booking)
		if resp != nil {
			responses[i] = *resp
		}
	}
	return responses
}
