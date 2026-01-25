package handler

import (
	"encoding/json"
	"net/http"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/delivery/http/middleware"
	"go-template-clean-architecture/internal/usecase"
	"go-template-clean-architecture/pkg/jwt"
	"go-template-clean-architecture/pkg/response"
	"go-template-clean-architecture/pkg/validator"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
	validator   *validator.CustomValidator
	jwtService  *jwt.JWTService
}

func NewAuthHandler(authUsecase usecase.AuthUsecase, validator *validator.CustomValidator, jwtService *jwt.JWTService) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		validator:   validator,
		jwtService:  jwtService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with email, password, and name
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Request"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	user, err := h.authUsecase.Register(r.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrEmailAlreadyExists:
			response.Error(w, http.StatusConflict, "Email already exists", nil)
		default:
			response.InternalServerError(w, "Failed to register user")
		}
		return
	}

	response.Success(w, http.StatusCreated, "User registered successfully", user)
}

// Login handles user login
// @Summary Login user
// @Description Login with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	tokens, err := h.authUsecase.Login(r.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrInvalidCredentials:
			response.Error(w, http.StatusUnauthorized, "Invalid email or password", nil)
		default:
			response.InternalServerError(w, "Failed to login")
		}
		return
	}

	response.Success(w, http.StatusOK, "Login successful", tokens)
}

// Logout handles user logout
// @Summary Logout user
// @Description Logout and revoke tokens
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenID, ok := middleware.GetTokenIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Invalid token")
		return
	}

	// Get refresh token from request body if provided
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	refreshTokenID := ""
	if req.RefreshToken != "" {
		claims, err := h.jwtService.ValidateToken(req.RefreshToken)
		if err == nil {
			refreshTokenID = claims.TokenID
		}
	}

	if err := h.authUsecase.Logout(r.Context(), tokenID, refreshTokenID); err != nil {
		response.InternalServerError(w, "Failed to logout")
		return
	}

	response.Success(w, http.StatusOK, "Logout successful", nil)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, h.validator.FormatValidationErrors(err))
		return
	}

	tokens, err := h.authUsecase.RefreshToken(r.Context(), &req)
	if err != nil {
		switch err {
		case usecase.ErrInvalidToken, usecase.ErrTokenRevoked:
			response.Error(w, http.StatusUnauthorized, err.Error(), nil)
		default:
			response.InternalServerError(w, "Failed to refresh token")
		}
		return
	}

	response.Success(w, http.StatusOK, "Token refreshed successfully", tokens)
}

// GetCurrentUser handles getting current user info
// @Summary Get current user
// @Description Get authenticated user information
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "Invalid token")
		return
	}

	user, err := h.authUsecase.GetCurrentUser(r.Context(), userID)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			response.NotFound(w, "User not found")
		default:
			response.InternalServerError(w, "Failed to get user info")
		}
		return
	}

	response.Success(w, http.StatusOK, "User retrieved successfully", user)
}
