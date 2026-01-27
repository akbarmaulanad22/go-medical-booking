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

// UserWithDoctorProfileToResponse converts a User entity (with DoctorProfile) to DoctorResponse DTO
func UserWithDoctorProfileToResponse(user *entity.User) *dto.DoctorResponse {
	if user == nil || user.DoctorProfile == nil {
		return nil
	}

	return &dto.DoctorResponse{
		ID:             user.ID,
		Email:          user.Email,
		FullName:       user.FullName,
		STRNumber:      user.DoctorProfile.STRNumber,
		Specialization: user.DoctorProfile.Specialization,
		Biography:      user.DoctorProfile.Biography,
		IsActive:       user.IsActive,
	}
}

// BuildDoctorResponse builds DoctorResponse from separate User and DoctorProfile entities
func BuildDoctorResponse(user *entity.User, profile *entity.DoctorProfile) *dto.DoctorResponse {
	if user == nil || profile == nil {
		return nil
	}

	return &dto.DoctorResponse{
		ID:             user.ID,
		Email:          user.Email,
		FullName:       user.FullName,
		STRNumber:      profile.STRNumber,
		Specialization: profile.Specialization,
		Biography:      profile.Biography,
		IsActive:       user.IsActive,
	}
}
