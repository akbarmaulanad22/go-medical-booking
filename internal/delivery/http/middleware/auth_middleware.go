package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go-template-clean-architecture/pkg/jwt"
	"go-template-clean-architecture/pkg/response"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	RoleIDKey    contextKey = "role_id"
	TokenIDKey   contextKey = "token_id"
)

type AuthMiddleware struct {
	jwtService  *jwt.JWTService
	redisClient *redis.Client
}

func NewAuthMiddleware(jwtService *jwt.JWTService, redisClient *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:  jwtService,
		redisClient: redisClient,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(w, "Authorization header is required")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(w, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate JWT token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(w, "Invalid or expired token")
			return
		}

		// Check if it's an access token
		if claims.TokenType != jwt.AccessToken {
			response.Unauthorized(w, "Invalid token type")
			return
		}

		// Check if token exists in Redis (not revoked)
		tokenKey := fmt.Sprintf("access_token:%s:%s", claims.UserID.String(), claims.TokenID)
		exists, err := m.redisClient.Exists(r.Context(), tokenKey).Result()
		if err != nil {
			response.InternalServerError(w, "Failed to validate token")
			return
		}
		if exists == 0 {
			response.Unauthorized(w, "Token has been revoked")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, RoleIDKey, claims.RoleID)
		ctx = context.WithValue(ctx, TokenIDKey, claims.TokenID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetUserEmailFromContext extracts user email from context
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetTokenIDFromContext extracts token ID from context
func GetTokenIDFromContext(ctx context.Context) (string, bool) {
	tokenID, ok := ctx.Value(TokenIDKey).(string)
	return tokenID, ok
}

// GetRoleIDFromContext extracts role ID from context
func GetRoleIDFromContext(ctx context.Context) (int, bool) {
	roleID, ok := ctx.Value(RoleIDKey).(int)
	return roleID, ok
}
