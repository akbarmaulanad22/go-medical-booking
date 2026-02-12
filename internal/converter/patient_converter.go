package converter

import (
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
)

// PatientProfileToResponse converts a PatientProfile entity + User entity to PatientResponse DTO
func PatientProfileToResponse(profile *entity.PatientProfile, user *entity.User) *dto.PatientResponse {
	if profile == nil || user == nil {
		return nil
	}

	return &dto.PatientResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		NIK:         profile.NIK,
		PhoneNumber: profile.PhoneNumber,
		DateOfBirth: profile.DateOfBirth.Format("2006-01-02"),
		Gender:      profile.Gender,
		Address:     profile.Address,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
