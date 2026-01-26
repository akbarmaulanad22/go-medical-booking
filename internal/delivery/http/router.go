package http

import (
	"net/http"

	"go-template-clean-architecture/internal/delivery/http/handler"
	"go-template-clean-architecture/internal/delivery/http/middleware"

	"github.com/gorilla/mux"
)

type Router struct {
	router         *mux.Router
	authHandler    *handler.AuthHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(
	authHandler *handler.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		router:         mux.NewRouter(),
		authHandler:    authHandler,
		authMiddleware: authMiddleware,
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

	// Auth routes (protected)
	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(r.authMiddleware.Authenticate)
	authProtected.HandleFunc("/logout", r.authHandler.Logout).Methods(http.MethodPost)
	authProtected.HandleFunc("/me", r.authHandler.GetCurrentUser).Methods(http.MethodGet)

	// Add CORS middleware
	r.router.Use(r.corsMiddleware)

	return r.router
}

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, req)
	})
}
