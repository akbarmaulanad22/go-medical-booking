package handler

import (
	"encoding/json"
	"net/http"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/service"
	"go-template-clean-architecture/internal/usecase"
	"go-template-clean-architecture/pkg/response"
	"go-template-clean-architecture/pkg/validator"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type BookingHandler struct {
	bookingUsecase usecase.PatientBookingUsecase
	validator      *validator.CustomValidator
}

func NewBookingHandler(bookingUsecase usecase.PatientBookingUsecase, validator *validator.CustomValidator) *BookingHandler {
	return &BookingHandler{
		bookingUsecase: bookingUsecase,
		validator:      validator,
	}
}

func (h *BookingHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.bookingUsecase.GetMyBookings(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to get bookings")
		return
	}

	response.Success(w, http.StatusOK, "Bookings retrieved successfully", bookings)
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	booking, err := h.bookingUsecase.CreateBooking(r.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrScheduleNotFound:
			response.NotFound(w, "Schedule not found")
		case usecase.ErrSchedulePast:
			response.Error(w, http.StatusBadRequest, "Cannot book a past schedule", nil)
		case usecase.ErrAlreadyBooked:
			response.Error(w, http.StatusConflict, "You have already booked this schedule", nil)
		case service.ErrQuotaFull:
			response.Error(w, http.StatusConflict, "Schedule slot is full, no remaining quota", nil)
		default:
			response.InternalServerError(w, "Failed to create booking")
		}
		return
	}

	response.Success(w, http.StatusCreated, "Booking created successfully", booking)
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookingID, err := uuid.Parse(vars["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid booking ID", nil)
		return
	}

	err = h.bookingUsecase.CancelBooking(r.Context(), bookingID)
	if err != nil {
		switch err {
		case usecase.ErrBookingNotFound:
			response.NotFound(w, "Booking not found")
		case usecase.ErrBookingNotOwned:
			response.Forbidden(w, "Booking does not belong to you")
		case usecase.ErrBookingAlreadyCancelled:
			response.Error(w, http.StatusConflict, "Booking is already cancelled", nil)
		default:
			response.InternalServerError(w, "Failed to cancel booking")
		}
		return
	}

	response.Success(w, http.StatusOK, "Booking cancelled successfully", nil)
}
