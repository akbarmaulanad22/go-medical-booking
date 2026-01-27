package converter

import (
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
)

// DoctorProfileToResponse converts a DoctorProfile entity to DoctorResponse DTO
func DoctorProfileToResponse(profile *entity.DoctorProfile) *dto.DoctorResponse {
	if profile == nil {
		return nil
	}

	return &dto.DoctorResponse{
		ID:             profile.UserID,
		Email:          profile.User.Email,
		FullName:       profile.User.FullName,
		STRNumber:      profile.STRNumber,
		Specialization: profile.Specialization,
		Biography:      profile.Biography,
		IsActive:       profile.User.IsActive,
	}
}

// DoctorProfilesToResponses converts a slice of DoctorProfile entities to slice of DoctorResponse DTOs
func DoctorProfilesToResponses(profiles []entity.DoctorProfile) []dto.DoctorResponse {
	responses := make([]dto.DoctorResponse, len(profiles))
	for i, profile := range profiles {
		responses[i] = dto.DoctorResponse{
			ID:             profile.UserID,
			Email:          profile.User.Email,
			FullName:       profile.User.FullName,
			STRNumber:      profile.STRNumber,
			Specialization: profile.Specialization,
			Biography:      profile.Biography,
			IsActive:       profile.User.IsActive,
		}
	}
	return responses
}
