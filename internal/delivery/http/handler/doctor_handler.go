package handler

import (
	"encoding/json"
	"net/http"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/delivery/http/middleware"
	"go-template-clean-architecture/internal/usecase"
	"go-template-clean-architecture/pkg/response"
	"go-template-clean-architecture/pkg/validator"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type DoctorHandler struct {
	doctorUsecase usecase.DoctorProfileUsecase
	validator     *validator.CustomValidator
}

func NewDoctorHandler(doctorUsecase usecase.DoctorProfileUsecase, validator *validator.CustomValidator) *DoctorHandler {
	return &DoctorHandler{
		doctorUsecase: doctorUsecase,
		validator:     validator,
	}
}

func (h *DoctorHandler) CreateDoctor(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDoctorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	doctor, err := h.doctorUsecase.CreateDoctor(r.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrDoctorEmailExists:
			response.Error(w, http.StatusConflict, "Email already exists", nil)
		case usecase.ErrDoctorSTRExists:
			response.Error(w, http.StatusConflict, "STR number already exists", nil)
		case usecase.ErrDoctorRoleNotFound:
			response.Error(w, http.StatusBadRequest, "Role not found", nil)
		default:
			response.InternalServerError(w, "Failed to create doctor")
		}
		return
	}

	response.Success(w, http.StatusCreated, "Doctor created successfully", doctor)
}

func (h *DoctorHandler) GetDoctor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	doctorID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid doctor ID", nil)
		return
	}

	doctor, err := h.doctorUsecase.GetDoctor(r.Context(), doctorID)
	if err != nil {
		if err == usecase.ErrDoctorNotFound {
			response.NotFound(w, "Doctor not found")
			return
		}
		response.InternalServerError(w, "Failed to get doctor")
		return
	}

	response.Success(w, http.StatusOK, "Doctor retrieved successfully", doctor)
}

func (h *DoctorHandler) GetAllDoctors(w http.ResponseWriter, r *http.Request) {
	doctors, err := h.doctorUsecase.GetAllDoctors(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get doctors")
		return
	}

	response.Success(w, http.StatusOK, "Doctors retrieved successfully", doctors)
}

func (h *DoctorHandler) UpdateDoctor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	doctorID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid doctor ID", nil)
		return
	}

	var req dto.UpdateDoctorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	doctor, err := h.doctorUsecase.UpdateDoctor(r.Context(), doctorID, &req)
	if err != nil {
		switch err {
		case usecase.ErrDoctorNotFound:
			response.NotFound(w, "Doctor not found")
		case usecase.ErrDoctorSTRExists:
			response.Error(w, http.StatusConflict, "STR number already exists", nil)
		default:
			response.InternalServerError(w, "Failed to update doctor")
		}
		return
	}

	response.Success(w, http.StatusOK, "Doctor updated successfully", doctor)
}

func (h *DoctorHandler) DeleteDoctor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	doctorID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid doctor ID", nil)
		return
	}

	err = h.doctorUsecase.DeleteDoctor(r.Context(), doctorID)
	if err != nil {
		if err == usecase.ErrDoctorNotFound {
			response.NotFound(w, "Doctor not found")
			return
		}
		response.InternalServerError(w, "Failed to delete doctor")
		return
	}

	response.Success(w, http.StatusOK, "Doctor deleted successfully", nil)
}

func (h *DoctorHandler) UpdateSelfProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	var req dto.DoctorUpdateSelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	doctor, err := h.doctorUsecase.UpdateSelfProfile(r.Context(), userID, &req)
	if err != nil {
		switch err {
		case usecase.ErrDoctorNotFound:
			response.NotFound(w, "Doctor not found")
		case usecase.ErrInvalidOldPassword:
			response.Error(w, http.StatusBadRequest, "Invalid old password", nil)
		default:
			response.InternalServerError(w, "Failed to update profile")
		}
		return
	}

	response.Success(w, http.StatusOK, "Profile updated successfully", doctor)
}
