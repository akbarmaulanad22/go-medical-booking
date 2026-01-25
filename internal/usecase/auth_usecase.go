package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"
	"go-template-clean-architecture/pkg/jwt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthUsecase interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error)
	Logout(ctx context.Context, accessTokenID, refreshTokenID string) error
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}

type authUsecase struct {
	userRepo    repository.UserRepository
	jwtService  *jwt.JWTService
	redisClient *redis.Client
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	jwtService *jwt.JWTService,
	redisClient *redis.Client,
) AuthUsecase {
	return &authUsecase{
		userRepo:    userRepo,
		jwtService:  jwtService,
		redisClient: redisClient,
	}
}

func (u *authUsecase) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	// Check if email already exists
	existingUser, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &entity.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (u *authUsecase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error) {
	// Find user by email
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, accessTokenID, err := u.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshTokenID, err := u.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Store tokens in Redis
	accessKey := fmt.Sprintf("access_token:%s:%s", user.ID.String(), accessTokenID)
	refreshKey := fmt.Sprintf("refresh_token:%s:%s", user.ID.String(), refreshTokenID)

	if err := u.redisClient.Set(ctx, accessKey, "valid", u.jwtService.GetAccessExpiry()).Err(); err != nil {
		return nil, err
	}

	if err := u.redisClient.Set(ctx, refreshKey, "valid", u.jwtService.GetRefreshExpiry()).Err(); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(u.jwtService.GetAccessExpiry().Seconds()),
	}, nil
}

func (u *authUsecase) Logout(ctx context.Context, accessTokenID, refreshTokenID string) error {
	// Delete tokens from Redis (pattern matching to find and delete)
	accessPattern := fmt.Sprintf("access_token:*:%s", accessTokenID)
	refreshPattern := fmt.Sprintf("refresh_token:*:%s", refreshTokenID)

	// Delete access token
	accessKeys, err := u.redisClient.Keys(ctx, accessPattern).Result()
	if err != nil {
		return err
	}
	if len(accessKeys) > 0 {
		if err := u.redisClient.Del(ctx, accessKeys...).Err(); err != nil {
			return err
		}
	}

	// Delete refresh token
	refreshKeys, err := u.redisClient.Keys(ctx, refreshPattern).Result()
	if err != nil {
		return err
	}
	if len(refreshKeys) > 0 {
		if err := u.redisClient.Del(ctx, refreshKeys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (u *authUsecase) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error) {
	// Validate refresh token
	claims, err := u.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != jwt.RefreshToken {
		return nil, ErrInvalidToken
	}

	// Check if refresh token exists in Redis
	refreshKey := fmt.Sprintf("refresh_token:%s:%s", claims.UserID.String(), claims.TokenID)
	exists, err := u.redisClient.Exists(ctx, refreshKey).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, ErrTokenRevoked
	}

	// Delete old refresh token
	if err := u.redisClient.Del(ctx, refreshKey).Err(); err != nil {
		return nil, err
	}

	// Generate new tokens
	accessToken, accessTokenID, err := u.jwtService.GenerateAccessToken(claims.UserID, claims.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshTokenID, err := u.jwtService.GenerateRefreshToken(claims.UserID, claims.Email)
	if err != nil {
		return nil, err
	}

	// Store new tokens in Redis
	accessKeyNew := fmt.Sprintf("access_token:%s:%s", claims.UserID.String(), accessTokenID)
	refreshKeyNew := fmt.Sprintf("refresh_token:%s:%s", claims.UserID.String(), refreshTokenID)

	if err := u.redisClient.Set(ctx, accessKeyNew, "valid", u.jwtService.GetAccessExpiry()).Err(); err != nil {
		return nil, err
	}

	if err := u.redisClient.Set(ctx, refreshKeyNew, "valid", u.jwtService.GetRefreshExpiry()).Err(); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(u.jwtService.GetAccessExpiry().Seconds()),
	}, nil
}

func (u *authUsecase) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Helper function to check if token is valid in Redis
func (u *authUsecase) IsTokenValid(ctx context.Context, userID uuid.UUID, tokenID string, tokenType jwt.TokenType) (bool, error) {
	var key string
	if tokenType == jwt.AccessToken {
		key = fmt.Sprintf("access_token:%s:%s", userID.String(), tokenID)
	} else {
		key = fmt.Sprintf("refresh_token:%s:%s", userID.String(), tokenID)
	}

	exists, err := u.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// RevokeAllUserTokens revokes all tokens for a user (useful when password changed or account compromised)
func (u *authUsecase) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	// Delete all access tokens for user
	accessPattern := fmt.Sprintf("access_token:%s:*", userID.String())
	accessKeys, err := u.redisClient.Keys(ctx, accessPattern).Result()
	if err != nil {
		return err
	}
	if len(accessKeys) > 0 {
		if err := u.redisClient.Del(ctx, accessKeys...).Err(); err != nil {
			return err
		}
	}

	// Delete all refresh tokens for user
	refreshPattern := fmt.Sprintf("refresh_token:%s:*", userID.String())
	refreshKeys, err := u.redisClient.Keys(ctx, refreshPattern).Result()
	if err != nil {
		return err
	}
	if len(refreshKeys) > 0 {
		if err := u.redisClient.Del(ctx, refreshKeys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

// Compile time check
var _ time.Duration
