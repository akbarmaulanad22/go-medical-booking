package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/usecase"
	"go-template-clean-architecture/pkg/response"
	"go-template-clean-architecture/pkg/validator"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type DoctorScheduleHandler struct {
	scheduleUsecase usecase.DoctorScheduleUsecase
	validator       *validator.CustomValidator
}

func NewDoctorScheduleHandler(scheduleUsecase usecase.DoctorScheduleUsecase, validator *validator.CustomValidator) *DoctorScheduleHandler {
	return &DoctorScheduleHandler{
		scheduleUsecase: scheduleUsecase,
		validator:       validator,
	}
}

func (h *DoctorScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	schedule, err := h.scheduleUsecase.CreateSchedule(r.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrDoctorNotFound:
			response.NotFound(w, "Doctor not found")
		case usecase.ErrInvalidScheduleDate:
			response.Error(w, http.StatusBadRequest, "Invalid schedule date format, use YYYY-MM-DD", nil)
		case usecase.ErrInvalidTimeFormat:
			response.Error(w, http.StatusBadRequest, "Invalid time format, use HH:MM", nil)
		default:
			response.InternalServerError(w, "Failed to create schedule")
		}
		return
	}

	response.Success(w, http.StatusCreated, "Schedule created successfully", schedule)
}

func (h *DoctorScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduleID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid schedule ID", nil)
		return
	}

	schedule, err := h.scheduleUsecase.GetSchedule(r.Context(), scheduleID)
	if err != nil {
		if err == usecase.ErrScheduleNotFound {
			response.NotFound(w, "Schedule not found")
			return
		}
		response.InternalServerError(w, "Failed to get schedule")
		return
	}

	response.Success(w, http.StatusOK, "Schedule retrieved successfully", schedule)
}

func (h *DoctorScheduleHandler) GetAllSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.scheduleUsecase.GetAllSchedules(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get schedules")
		return
	}

	response.Success(w, http.StatusOK, "Schedules retrieved successfully", schedules)
}

func (h *DoctorScheduleHandler) GetSchedulesByDoctor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	doctorID, err := uuid.Parse(vars["doctorId"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid doctor ID", nil)
		return
	}

	schedules, err := h.scheduleUsecase.GetSchedulesByDoctor(r.Context(), doctorID)
	if err != nil {
		response.InternalServerError(w, "Failed to get schedules")
		return
	}

	response.Success(w, http.StatusOK, "Schedules retrieved successfully", schedules)
}

func (h *DoctorScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduleID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid schedule ID", nil)
		return
	}

	var req dto.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	schedule, err := h.scheduleUsecase.UpdateSchedule(r.Context(), scheduleID, &req)
	if err != nil {
		switch err {
		case usecase.ErrScheduleNotFound:
			response.NotFound(w, "Schedule not found")
		case usecase.ErrDoctorNotFound:
			response.NotFound(w, "Doctor not found")
		case usecase.ErrInvalidScheduleDate:
			response.Error(w, http.StatusBadRequest, "Invalid schedule date format, use YYYY-MM-DD", nil)
		case usecase.ErrInvalidTimeFormat:
			response.Error(w, http.StatusBadRequest, "Invalid time format, use HH:MM", nil)
		default:
			response.InternalServerError(w, "Failed to update schedule")
		}
		return
	}

	response.Success(w, http.StatusOK, "Schedule updated successfully", schedule)
}

func (h *DoctorScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	scheduleID, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid schedule ID", nil)
		return
	}

	err = h.scheduleUsecase.DeleteSchedule(r.Context(), scheduleID)
	if err != nil {
		if err == usecase.ErrScheduleNotFound {
			response.NotFound(w, "Schedule not found")
			return
		}
		response.InternalServerError(w, "Failed to delete schedule")
		return
	}

	response.Success(w, http.StatusOK, "Schedule deleted successfully", nil)
}
