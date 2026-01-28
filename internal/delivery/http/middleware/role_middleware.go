package middleware

import (
	"net/http"

	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/pkg/response"
)

// RequireRole creates a middleware that checks if the user has any of the required roles
// Role is read from context (set by AuthMiddleware from JWT claims)
func RequireRole(allowedRoleIDs ...int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role ID from context (set by AuthMiddleware)
			roleID, ok := GetRoleIDFromContext(r.Context())
			if !ok {
				response.Unauthorized(w, "Role information not found")
				return
			}

			// Check if user's role is in allowed roles
			allowed := false
			for _, allowedRoleID := range allowedRoleIDs {
				if roleID == allowedRoleID {
					allowed = true
					break
				}
			}

			if !allowed {
				response.Forbidden(w, "You don't have permission to access this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a convenience middleware for admin-only endpoints
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(entity.RoleIDAdmin)(next)
}

// RequireDoctor is a convenience middleware for doctor-only endpoints
func RequireDoctor(next http.Handler) http.Handler {
	return RequireRole(entity.RoleIDDoctor)(next)
}

// RequirePatient is a convenience middleware for patient-only endpoints
func RequirePatient(next http.Handler) http.Handler {
	return RequireRole(entity.RoleIDPatient)(next)
}

// RequireAdminOrDoctor is a convenience middleware for admin or doctor endpoints
func RequireAdminOrDoctor(next http.Handler) http.Handler {
	return RequireRole(entity.RoleIDAdmin, entity.RoleIDDoctor)(next)
}
