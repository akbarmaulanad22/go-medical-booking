package converter

import (
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"

	"github.com/google/uuid"
)

// ScheduleToResponse converts a DoctorSchedule entity to ScheduleResponse DTO
func ScheduleToResponse(schedule *entity.DoctorSchedule) *dto.ScheduleResponse {
	if schedule == nil {
		return nil
	}

	response := &dto.ScheduleResponse{
		ID:           schedule.ID,
		DoctorID:     schedule.DoctorID,
		ScheduleDate: schedule.ScheduleDate.Format("2006-01-02"),
		StartTime:    schedule.StartTime,
		EndTime:      schedule.EndTime,
		TotalQuota:   schedule.TotalQuota,
		CreatedAt:    schedule.CreatedAt,
		UpdatedAt:    schedule.UpdatedAt,
	}

	// Include doctor info if available
	if schedule.Doctor.UserID != uuid.Nil {
		response.Doctor = DoctorProfileToResponse(&schedule.Doctor)
	}

	return response
}

// SchedulesToResponses converts a slice of DoctorSchedule entities to slice of ScheduleResponse DTOs
func SchedulesToResponses(schedules []entity.DoctorSchedule) []dto.ScheduleResponse {
	responses := make([]dto.ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		response := dto.ScheduleResponse{
			ID:           schedule.ID,
			DoctorID:     schedule.DoctorID,
			ScheduleDate: schedule.ScheduleDate.Format("2006-01-02"),
			StartTime:    schedule.StartTime,
			EndTime:      schedule.EndTime,
			TotalQuota:   schedule.TotalQuota,
			CreatedAt:    schedule.CreatedAt,
			UpdatedAt:    schedule.UpdatedAt,
		}

		// Include doctor info if available
		if schedule.Doctor.UserID != uuid.Nil {
			response.Doctor = DoctorProfileToResponse(&schedule.Doctor)
		}

		responses[i] = response
	}
	return responses
}
