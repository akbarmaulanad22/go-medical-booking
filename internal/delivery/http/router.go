package http

import (
	"net/http"

	"go-template-clean-architecture/internal/delivery/http/handler"
	"go-template-clean-architecture/internal/delivery/http/middleware"

	"github.com/gorilla/mux"
)

type Router struct {
	router                *mux.Router
	authHandler           *handler.AuthHandler
	doctorHandler         *handler.DoctorHandler
	doctorScheduleHandler *handler.DoctorScheduleHandler
	bookingHandler        *handler.BookingHandler
	authMiddleware        *middleware.AuthMiddleware
	corsMiddleware        *middleware.CORSMiddleware
	auditHandler          *handler.AuditLogHandler
}

func NewRouter(
	authHandler *handler.AuthHandler,
	doctorHandler *handler.DoctorHandler,
	doctorScheduleHandler *handler.DoctorScheduleHandler,
	bookingHandler *handler.BookingHandler,
	authMiddleware *middleware.AuthMiddleware,
	corsMiddleware *middleware.CORSMiddleware,
	auditHandler *handler.AuditLogHandler,
) *Router {
	return &Router{
		router:                mux.NewRouter(),
		authHandler:           authHandler,
		doctorHandler:         doctorHandler,
		doctorScheduleHandler: doctorScheduleHandler,
		bookingHandler:        bookingHandler,
		authMiddleware:        authMiddleware,
		corsMiddleware:        corsMiddleware,
		auditHandler:          auditHandler,
	}
}

func (r *Router) Setup() *mux.Router {
	// API versioning
	api := r.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", r.healthCheck).Methods(http.MethodGet)

	// Auth routes (public)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register/patient", r.authHandler.RegisterPatient).Methods(http.MethodPost)
	auth.HandleFunc("/register/doctor", r.authHandler.RegisterDoctor).Methods(http.MethodPost)
	auth.HandleFunc("/login", r.authHandler.Login).Methods(http.MethodPost)
	auth.HandleFunc("/refresh-token", r.authHandler.RefreshToken).Methods(http.MethodPost)

	// Public routes
	public := api.PathPrefix("/").Subrouter()
	public.HandleFunc("/doctors", r.doctorHandler.GetAllDoctors).Methods(http.MethodGet)
	// public.HandleFunc("/doctors/{id}", r.doctorHandler.GetDoctor).Methods(http.MethodGet)
	public.HandleFunc("/schedules", r.doctorScheduleHandler.GetAllSchedules).Methods(http.MethodGet)
	// public.HandleFunc("/schedules/{id}", r.doctorScheduleHandler.GetSchedule).Methods(http.MethodGet)

	// Auth routes (protected)
	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(r.authMiddleware.Authenticate)
	authProtected.HandleFunc("/logout", r.authHandler.Logout).Methods(http.MethodPost)
	authProtected.HandleFunc("/me", r.authHandler.GetCurrentUser).Methods(http.MethodGet)

	// Admin routes (protected - admin only)
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(r.authMiddleware.Authenticate)
	admin.Use(middleware.RequireAdmin)

	// Doctor management (admin)
	admin.HandleFunc("/doctors", r.doctorHandler.CreateDoctor).Methods(http.MethodPost)
	admin.HandleFunc("/doctors", r.doctorHandler.GetAllDoctors).Methods(http.MethodGet)
	admin.HandleFunc("/doctors/{id}", r.doctorHandler.GetDoctor).Methods(http.MethodGet)
	admin.HandleFunc("/doctors/{id}", r.doctorHandler.UpdateDoctor).Methods(http.MethodPut)
	admin.HandleFunc("/doctors/{id}", r.doctorHandler.DeleteDoctor).Methods(http.MethodDelete)

	// Schedule management (admin)
	admin.HandleFunc("/schedules", r.doctorScheduleHandler.CreateSchedule).Methods(http.MethodPost)
	admin.HandleFunc("/schedules", r.doctorScheduleHandler.GetAllSchedules).Methods(http.MethodGet)
	admin.HandleFunc("/schedules/{id}", r.doctorScheduleHandler.GetSchedule).Methods(http.MethodGet)
	admin.HandleFunc("/schedules/{id}", r.doctorScheduleHandler.UpdateSchedule).Methods(http.MethodPut)
	admin.HandleFunc("/schedules/{id}", r.doctorScheduleHandler.DeleteSchedule).Methods(http.MethodDelete)
	admin.HandleFunc("/doctors/{doctorId}/schedules", r.doctorScheduleHandler.GetSchedulesByDoctor).Methods(http.MethodGet)

	// Audit Log
	admin.HandleFunc("/audit-logs", r.auditHandler.GetAllAuditLogs).Methods(http.MethodGet)
	admin.HandleFunc("/audit-logs/{id}", r.auditHandler.GetAuditLog).Methods(http.MethodGet)

	// Doctor routes (protected - doctor only)
	doctor := api.PathPrefix("/doctor").Subrouter()
	doctor.Use(r.authMiddleware.Authenticate)
	doctor.Use(middleware.RequireDoctor)
	doctor.HandleFunc("/schedules", r.doctorScheduleHandler.GetMySchedules).Methods(http.MethodGet)
	doctor.HandleFunc("/profile", r.doctorHandler.UpdateSelfProfile).Methods(http.MethodPut)

	// Patient routes (protected - patient only)
	patient := api.PathPrefix("/patient").Subrouter()
	patient.Use(r.authMiddleware.Authenticate)
	patient.Use(middleware.RequirePatient)
	patient.HandleFunc("/bookings", r.bookingHandler.GetMyBookings).Methods(http.MethodGet)
	patient.HandleFunc("/bookings", r.bookingHandler.CreateBooking).Methods(http.MethodPost)
	patient.HandleFunc("/bookings/{id}/cancel", r.bookingHandler.CancelBooking).Methods(http.MethodPut)

	// Add CORS middleware
	r.router.Use(r.corsMiddleware.Handle)

	return r.router
}

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}
