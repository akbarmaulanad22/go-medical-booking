package handler

import (
	"encoding/json"
	"net/http"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/usecase"
	"go-template-clean-architecture/pkg/response"
	"go-template-clean-architecture/pkg/validator"
)

type PatientHandler struct {
	patientUsecase usecase.PatientProfileUsecase
	validator      *validator.CustomValidator
}

func NewPatientHandler(patientUsecase usecase.PatientProfileUsecase, validator *validator.CustomValidator) *PatientHandler {
	return &PatientHandler{
		patientUsecase: patientUsecase,
		validator:      validator,
	}
}

func (h *PatientHandler) UpdateSelfProfile(w http.ResponseWriter, r *http.Request) {
	var req dto.PatientUpdateSelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	profile, err := h.patientUsecase.UpdateSelfProfile(r.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrPatientNotFound:
			response.NotFound(w, "Patient profile not found")
		case usecase.ErrInvalidOldPassword:
			response.Error(w, http.StatusBadRequest, "Invalid old password", nil)
		default:
			response.InternalServerError(w, "Failed to update profile")
		}
		return
	}

	response.Success(w, http.StatusOK, "Profile updated successfully", profile)
}
