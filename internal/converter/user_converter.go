package converter

import (
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
)

// UserToResponse converts a User entity to UserResponse DTO
// Includes DoctorProfile and PatientProfile if they are loaded
func UserToResponse(user *entity.User) *dto.UserResponse {
	if user == nil {
		return nil
	}

	response := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role.RoleName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Include DoctorProfile if exists
	if user.DoctorProfile != nil {
		response.DoctorProfile = &dto.DoctorProfileResponse{
			STRNumber:      user.DoctorProfile.STRNumber,
			Specialization: user.DoctorProfile.Specialization,
			Biography:      user.DoctorProfile.Biography,
		}
	}

	// Include PatientProfile if exists
	if user.PatientProfile != nil {
		response.PatientProfile = &dto.PatientProfileResponse{
			UserID:      user.PatientProfile.UserID,
			NIK:         user.PatientProfile.NIK,
			PhoneNumber: user.PatientProfile.PhoneNumber,
			DateOfBirth: user.PatientProfile.DateOfBirth.Format("2006-01-02"),
			Gender:      user.PatientProfile.Gender,
			Address:     user.PatientProfile.Address,
		}
	}

	return response
}

// UserToResponseWithRole converts a User entity to UserResponse DTO with explicit role name
// Use this when Role is not preloaded but roleID is known
// func UserToResponseWithRole(user *entity.User, roleName string) *dto.UserResponse {
// 	if user == nil {
// 		return nil
// 	}

// 	response := &dto.UserResponse{
// 		ID:        user.ID,
// 		Email:     user.Email,
// 		FullName:  user.FullName,
// 		Role:      roleName,
// 		CreatedAt: user.CreatedAt,
// 		UpdatedAt: user.UpdatedAt,
// 	}

// 	// Include DoctorProfile if exists
// 	if user.DoctorProfile != nil {
// 		response.DoctorProfile = &dto.DoctorProfileResponse{
// 			STRNumber:      user.DoctorProfile.STRNumber,
// 			Specialization: user.DoctorProfile.Specialization,
// 			Biography:      user.DoctorProfile.Biography,
// 		}
// 	}

// 	// Include PatientProfile if exists
// 	if user.PatientProfile != nil {
// 		response.PatientProfile = &dto.PatientProfileResponse{
// 			UserID:      user.PatientProfile.UserID,
// 			NIK:         user.PatientProfile.NIK,
// 			PhoneNumber: user.PatientProfile.PhoneNumber,
// 			DateOfBirth: user.PatientProfile.DateOfBirth.Format("2006-01-02"),
// 			Gender:      user.PatientProfile.Gender,
// 			Address:     user.PatientProfile.Address,
// 		}
// 	}

// 	return response
// }
