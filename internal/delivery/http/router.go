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
	productHandler *handler.ProductHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewRouter(
	authHandler *handler.AuthHandler,
	productHandler *handler.ProductHandler,
	authMiddleware *middleware.AuthMiddleware,
) *Router {
	return &Router{
		router:         mux.NewRouter(),
		authHandler:    authHandler,
		productHandler: productHandler,
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
	auth.HandleFunc("/register", r.authHandler.Register).Methods(http.MethodPost)
	auth.HandleFunc("/login", r.authHandler.Login).Methods(http.MethodPost)
	auth.HandleFunc("/refresh-token", r.authHandler.RefreshToken).Methods(http.MethodPost)

	// Auth routes (protected)
	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(r.authMiddleware.Authenticate)
	authProtected.HandleFunc("/logout", r.authHandler.Logout).Methods(http.MethodPost)
	authProtected.HandleFunc("/me", r.authHandler.GetCurrentUser).Methods(http.MethodGet)

	// Product routes (public)
	products := api.PathPrefix("/products").Subrouter()
	products.HandleFunc("", r.productHandler.GetAll).Methods(http.MethodGet)
	products.HandleFunc("/{id}", r.productHandler.GetByID).Methods(http.MethodGet)

	// Product routes (protected)
	productsProtected := api.PathPrefix("/products").Subrouter()
	productsProtected.Use(r.authMiddleware.Authenticate)
	productsProtected.HandleFunc("", r.productHandler.Create).Methods(http.MethodPost)
	productsProtected.HandleFunc("/{id}", r.productHandler.Update).Methods(http.MethodPut)
	productsProtected.HandleFunc("/{id}", r.productHandler.Delete).Methods(http.MethodDelete)

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
