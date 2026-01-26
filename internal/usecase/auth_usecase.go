package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"
	"go-template-clean-architecture/pkg/jwt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrUserNotFound       = errors.New("user not found")
	ErrRoleNotFound       = errors.New("role not found")
	ErrNIKAlreadyExists   = errors.New("NIK already exists")
	ErrSTRAlreadyExists   = errors.New("STR number already exists")
	ErrInvalidDateFormat  = errors.New("invalid date format, use YYYY-MM-DD")
)

type AuthUsecase interface {
	RegisterPatient(ctx context.Context, req *dto.RegisterPatientRequest) (*dto.UserResponse, error)
	RegisterDoctor(ctx context.Context, req *dto.RegisterDoctorRequest) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error)
	Logout(ctx context.Context, accessTokenID, refreshTokenID string) error
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}

type authUsecase struct {
	db                 *gorm.DB
	log                *logrus.Logger
	userRepo           repository.UserRepository
	roleRepo           repository.RoleRepository
	doctorProfileRepo  repository.DoctorProfileRepository
	patientProfileRepo repository.PatientProfileRepository
	jwtService         *jwt.JWTService
	redisClient        *redis.Client
}

func NewAuthUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	doctorProfileRepo repository.DoctorProfileRepository,
	patientProfileRepo repository.PatientProfileRepository,
	jwtService *jwt.JWTService,
	redisClient *redis.Client,
) AuthUsecase {
	return &authUsecase{
		db:                 db,
		log:                log,
		userRepo:           userRepo,
		roleRepo:           roleRepo,
		doctorProfileRepo:  doctorProfileRepo,
		patientProfileRepo: patientProfileRepo,
		jwtService:         jwtService,
		redisClient:        redisClient,
	}
}

func (u *authUsecase) RegisterPatient(ctx context.Context, req *dto.RegisterPatientRequest) (*dto.UserResponse, error) {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Parse date of birth
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		u.log.Warnf("Failed to hash password: %+v", err)
		return nil, err
	}

	// Create user
	user := &entity.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		RoleID:   entity.RoleIDPatient,
		IsActive: true,
	}

	if err := u.userRepo.Create(tx, user); err != nil {
		if isDuplicateKeyError(err, "email") {
			return nil, ErrEmailAlreadyExists
		}
		if isForeignKeyError(err, "role") {
			return nil, ErrRoleNotFound
		}
		u.log.Warnf("Failed to create user: %+v", err)
		return nil, err
	}

	// Create patient profile
	patientProfile := &entity.PatientProfile{
		UserID:      user.ID,
		NIK:         req.NIK,
		PhoneNumber: req.PhoneNumber,
		DateOfBirth: dob,
		Gender:      req.Gender,
		Address:     req.Address,
	}

	if err := u.patientProfileRepo.Create(ctx, tx, patientProfile); err != nil {
		if isDuplicateKeyError(err, "nik") {
			return nil, ErrNIKAlreadyExists
		}
		u.log.Warnf("Failed to create patient profile: %+v", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      entity.RolePatient,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (u *authUsecase) RegisterDoctor(ctx context.Context, req *dto.RegisterDoctorRequest) (*dto.UserResponse, error) {
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		u.log.Warnf("Failed to hash password: %+v", err)
		return nil, err
	}

	// Create user
	user := &entity.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		FullName: req.FullName,
		RoleID:   entity.RoleIDDoctor,
		IsActive: true,
	}

	if err := u.userRepo.Create(tx, user); err != nil {
		if isDuplicateKeyError(err, "email") {
			return nil, ErrEmailAlreadyExists
		}
		if isForeignKeyError(err, "role") {
			return nil, ErrRoleNotFound
		}
		u.log.Warnf("Failed to create user: %+v", err)
		return nil, err
	}

	// Create doctor profile
	doctorProfile := &entity.DoctorProfile{
		UserID:         user.ID,
		STRNumber:      req.STRNumber,
		Specialization: req.Specialization,
		Biography:      req.Biography,
	}

	if err := u.doctorProfileRepo.Create(ctx, tx, doctorProfile); err != nil {
		if isDuplicateKeyError(err, "str_number") {
			return nil, ErrSTRAlreadyExists
		}
		u.log.Warnf("Failed to create doctor profile: %+v", err)
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		u.log.Warnf("Failed commit transaction: %+v", err)
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      entity.RoleDoctor,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (u *authUsecase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error) {
	// Find user by email (read-only, no transaction needed)
	user, err := u.userRepo.FindByEmail(u.db, req.Email)
	if err != nil {
		u.log.Warnf("Failed to find user by email: %+v", err)
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, accessTokenID, err := u.jwtService.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		u.log.Warnf("Failed to generate access token: %+v", err)
		return nil, err
	}

	refreshToken, refreshTokenID, err := u.jwtService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		u.log.Warnf("Failed to generate refresh token: %+v", err)
		return nil, err
	}

	// Store tokens in Redis
	accessKey := fmt.Sprintf("access_token:%s:%s", user.ID.String(), accessTokenID)
	refreshKey := fmt.Sprintf("refresh_token:%s:%s", user.ID.String(), refreshTokenID)

	if err := u.redisClient.Set(ctx, accessKey, "valid", u.jwtService.GetAccessExpiry()).Err(); err != nil {
		u.log.Warnf("Failed to store access token in Redis: %+v", err)
		return nil, err
	}

	if err := u.redisClient.Set(ctx, refreshKey, "valid", u.jwtService.GetRefreshExpiry()).Err(); err != nil {
		u.log.Warnf("Failed to store refresh token in Redis: %+v", err)
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
		u.log.Warnf("Failed to get access token keys: %+v", err)
		return err
	}
	if len(accessKeys) > 0 {
		if err := u.redisClient.Del(ctx, accessKeys...).Err(); err != nil {
			u.log.Warnf("Failed to delete access token: %+v", err)
			return err
		}
	}

	// Delete refresh token
	refreshKeys, err := u.redisClient.Keys(ctx, refreshPattern).Result()
	if err != nil {
		u.log.Warnf("Failed to get refresh token keys: %+v", err)
		return err
	}
	if len(refreshKeys) > 0 {
		if err := u.redisClient.Del(ctx, refreshKeys...).Err(); err != nil {
			u.log.Warnf("Failed to delete refresh token: %+v", err)
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
		u.log.Warnf("Failed to check refresh token in Redis: %+v", err)
		return nil, err
	}
	if exists == 0 {
		return nil, ErrTokenRevoked
	}

	// Delete old refresh token
	if err := u.redisClient.Del(ctx, refreshKey).Err(); err != nil {
		u.log.Warnf("Failed to delete old refresh token: %+v", err)
		return nil, err
	}

	// Generate new tokens
	accessToken, accessTokenID, err := u.jwtService.GenerateAccessToken(claims.UserID, claims.Email)
	if err != nil {
		u.log.Warnf("Failed to generate access token: %+v", err)
		return nil, err
	}

	refreshToken, refreshTokenID, err := u.jwtService.GenerateRefreshToken(claims.UserID, claims.Email)
	if err != nil {
		u.log.Warnf("Failed to generate refresh token: %+v", err)
		return nil, err
	}

	// Store new tokens in Redis
	accessKeyNew := fmt.Sprintf("access_token:%s:%s", claims.UserID.String(), accessTokenID)
	refreshKeyNew := fmt.Sprintf("refresh_token:%s:%s", claims.UserID.String(), refreshTokenID)

	if err := u.redisClient.Set(ctx, accessKeyNew, "valid", u.jwtService.GetAccessExpiry()).Err(); err != nil {
		u.log.Warnf("Failed to store access token in Redis: %+v", err)
		return nil, err
	}

	if err := u.redisClient.Set(ctx, refreshKeyNew, "valid", u.jwtService.GetRefreshExpiry()).Err(); err != nil {
		u.log.Warnf("Failed to store refresh token in Redis: %+v", err)
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(u.jwtService.GetAccessExpiry().Seconds()),
	}, nil
}

func (u *authUsecase) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(u.db, userID)
	if err != nil {
		u.log.Warnf("Failed to find user by ID: %+v", err)
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
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
		u.log.Warnf("Failed to check token validity: %+v", err)
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
		u.log.Warnf("Failed to get access token keys: %+v", err)
		return err
	}
	if len(accessKeys) > 0 {
		if err := u.redisClient.Del(ctx, accessKeys...).Err(); err != nil {
			u.log.Warnf("Failed to delete access tokens: %+v", err)
			return err
		}
	}

	// Delete all refresh tokens for user
	refreshPattern := fmt.Sprintf("refresh_token:%s:*", userID.String())
	refreshKeys, err := u.redisClient.Keys(ctx, refreshPattern).Result()
	if err != nil {
		u.log.Warnf("Failed to get refresh token keys: %+v", err)
		return err
	}
	if len(refreshKeys) > 0 {
		if err := u.redisClient.Del(ctx, refreshKeys...).Err(); err != nil {
			u.log.Warnf("Failed to delete refresh tokens: %+v", err)
			return err
		}
	}

	return nil
}

// isDuplicateKeyError checks if the error is a PostgreSQL unique constraint violation
// containing the specified constraint name
func isDuplicateKeyError(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL error code 23505 = unique_violation
		if pgErr.Code == "23505" && strings.Contains(strings.ToLower(pgErr.ConstraintName), strings.ToLower(constraintName)) {
			return true
		}
	}
	return false
}

// isForeignKeyError checks if the error is a PostgreSQL foreign key violation
// containing the specified constraint name
func isForeignKeyError(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL error code 23503 = foreign_key_violation
		if pgErr.Code == "23503" && strings.Contains(strings.ToLower(pgErr.ConstraintName), strings.ToLower(constraintName)) {
			return true
		}
	}
	return false
}

// Compile time check
var _ time.Duration
