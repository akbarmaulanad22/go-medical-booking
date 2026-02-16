package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-template-clean-architecture/internal/converter"
	"go-template-clean-architecture/internal/delivery/dto"
	"go-template-clean-architecture/internal/domain/entity"
	"go-template-clean-architecture/internal/domain/repository"
	"go-template-clean-architecture/internal/service"
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
	ErrAccountLocked      = errors.New("account temporarily locked, try again later")
)

// =============================================================================
// Constants
// =============================================================================

const (
	maxLoginAttempts    = 5
	loginLockoutPeriod  = 3 * time.Minute
	loginAttemptsPrefix = "login_attempts:"
)

// Lua script: atomically INCR attempt count and set TTL on first attempt
var loginRateLimitScript = redis.NewScript(`
	local current = redis.call('INCR', KEYS[1])
	if current == 1 then
		redis.call('EXPIRE', KEYS[1], ARGV[1])
	end
	return current
`)

// =============================================================================
// Interface & Struct
// =============================================================================

type AuthUsecase interface {
	Register(ctx context.Context, user *entity.User) (*dto.UserResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error)
	Logout(ctx context.Context, accessTokenID, refreshTokenID string) error
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}

type authUsecase struct {
	db           *gorm.DB
	log          *logrus.Logger
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	jwtService   *jwt.JWTService
	redisClient  *redis.Client
	auditService service.AuditService
}

func NewAuthUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	jwtService *jwt.JWTService,
	redisClient *redis.Client,
	auditService service.AuditService,
) AuthUsecase {
	return &authUsecase{
		db:           db,
		log:          log,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		jwtService:   jwtService,
		redisClient:  redisClient,
		auditService: auditService,
	}
}

// =============================================================================
// Register — unified for all roles
// =============================================================================

// Register creates a new user with any associated profile (PatientProfile / DoctorProfile).
// The handler is responsible for building the entity.User with the correct RoleID and
// profile relation attached. Password must be plaintext — this function hashes it.
//
// GORM auto-creates nested associations when the parent struct has them populated,
// so we only need a single db.Create(user) call.
func (u *authUsecase) Register(ctx context.Context, user *entity.User) (*dto.UserResponse, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		go u.log.Warnf("Failed to hash password: %+v", err)
		return nil, err
	}
	user.Password = string(hashedPassword)

	// Create user + associations in a transaction
	tx := u.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := u.userRepo.Create(tx, user); err != nil {
		go u.log.Warnf("Failed to create user: %+v", err)
		if isDuplicateKeyError(err, "email") {
			return nil, ErrEmailAlreadyExists
		}
		if isDuplicateKeyError(err, "nik") {
			return nil, ErrNIKAlreadyExists
		}
		if isDuplicateKeyError(err, "str_number") {
			return nil, ErrSTRAlreadyExists
		}
		if isForeignKeyError(err, "role") {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		go u.log.Warnf("Failed to commit transaction: %+v", err)
		return nil, err
	}

	// Non-blocking audit log: register success
	go func() {
		ctx := context.Background()
		if err := u.auditService.LogCreate(ctx, u.db, &user.ID, entity.AuditActionUserRegister, "user", user.ID.String(), entity.JSON{
			"email":   user.Email,
			"role_id": user.RoleID,
		}); err != nil {
			u.log.Warnf("Failed to log register audit: %+v", err)
		}
	}()

	return converter.UserToResponse(user), nil
}

// =============================================================================
// Login — with Redis rate limiting
// =============================================================================

func (u *authUsecase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error) {
	// ---- Rate Limit Check ----
	attemptsKey := fmt.Sprintf("%s%s", loginAttemptsPrefix, req.Email)

	count, err := u.redisClient.Get(ctx, attemptsKey).Int()
	if err != nil && !errors.Is(err, redis.Nil) {
		go u.log.Warnf("Failed to get login attempts: %+v", err)
		// Non-blocking: if Redis is down, allow login attempt
	}
	if count >= maxLoginAttempts {
		go u.log.Warnf("Account locked for email %s: too many login attempts", req.Email)
		// Non-blocking audit log: account locked
		go func() {
			ctx := context.Background()
			u.auditService.LogCreate(ctx, u.db, nil, "user.login_locked", "user", "", entity.JSON{
				"email":  req.Email,
				"reason": "too many login attempts",
			})
		}()
		return nil, ErrAccountLocked
	}

	// ---- Find User ----
	user, err := u.userRepo.FindByEmail(u.db, req.Email)
	if err != nil {
		go u.log.Warnf("Failed to find user by email: %+v", err)
		// Increment attempt on user-not-found to prevent enumeration
		u.incrementLoginAttempts(ctx, attemptsKey)
		return nil, ErrInvalidCredentials
	}

	// ---- Verify Password ----
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		go u.log.Warnf("Invalid credentials for email %s: %+v", req.Email, err)
		u.incrementLoginAttempts(ctx, attemptsKey)
		// Non-blocking audit log: login failed
		go func() {
			ctx := context.Background()
			u.auditService.LogCreate(ctx, u.db, &user.ID, "user.login_failed", "user", user.ID.String(), entity.JSON{
				"email":  req.Email,
				"reason": "invalid password",
			})
		}()
		return nil, ErrInvalidCredentials
	}

	// ---- Password correct: reset attempts ----
	if delErr := u.redisClient.Del(ctx, attemptsKey).Err(); delErr != nil {
		go u.log.Warnf("Failed to reset login attempts: %+v", delErr)
	}

	// ---- Generate Tokens ----
	accessToken, accessTokenID, err := u.jwtService.GenerateAccessToken(user.ID, user.Email, user.RoleID)
	if err != nil {
		go u.log.Warnf("Failed to generate access token: %+v", err)
		return nil, err
	}

	refreshToken, refreshTokenID, err := u.jwtService.GenerateRefreshToken(user.ID, user.Email, user.RoleID)
	if err != nil {
		go u.log.Warnf("Failed to generate refresh token: %+v", err)
		return nil, err
	}

	// ---- Store tokens in Redis ----
	accessKey := fmt.Sprintf("access_token:%s:%s", user.ID.String(), accessTokenID)
	refreshKey := fmt.Sprintf("refresh_token:%s:%s", user.ID.String(), refreshTokenID)

	if err := u.redisClient.Set(ctx, accessKey, "valid", u.jwtService.GetAccessExpiry()).Err(); err != nil {
		go u.log.Warnf("Failed to store access token in Redis: %+v", err)
		return nil, err
	}

	if err := u.redisClient.Set(ctx, refreshKey, "valid", u.jwtService.GetRefreshExpiry()).Err(); err != nil {
		go u.log.Warnf("Failed to store refresh token in Redis: %+v", err)
		return nil, err
	}

	// Non-blocking audit log: login success
	go func() {
		ctx := context.Background()
		u.auditService.LogCreate(ctx, u.db, &user.ID, entity.AuditActionUserLogin, "user", user.ID.String(), entity.JSON{
			"email": user.Email,
		})
	}()

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(u.jwtService.GetAccessExpiry().Seconds()),
	}, nil
}

// incrementLoginAttempts atomically increments the login attempt counter.
// Sets TTL to loginLockoutPeriod on first increment.
func (u *authUsecase) incrementLoginAttempts(ctx context.Context, key string) {
	lockoutSeconds := int(loginLockoutPeriod.Seconds())
	_, err := loginRateLimitScript.Run(ctx, u.redisClient, []string{key}, lockoutSeconds).Int()
	if err != nil {
		go u.log.Warnf("Failed to increment login attempts: %+v", err)
	}
}

// =============================================================================
// Logout
// =============================================================================

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

// =============================================================================
// RefreshToken
// =============================================================================

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
	accessToken, accessTokenID, err := u.jwtService.GenerateAccessToken(claims.UserID, claims.Email, claims.RoleID)
	if err != nil {
		u.log.Warnf("Failed to generate access token: %+v", err)
		return nil, err
	}

	refreshToken, refreshTokenID, err := u.jwtService.GenerateRefreshToken(claims.UserID, claims.Email, claims.RoleID)
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

// =============================================================================
// GetCurrentUser
// =============================================================================

func (u *authUsecase) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(u.db, userID)
	if err != nil {
		u.log.Warnf("Failed to find user by ID: %+v", err)
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return converter.UserToResponse(user), nil
}

// =============================================================================
// Helper: Token Validation
// =============================================================================

// IsTokenValid checks if a token is valid in Redis
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

// =============================================================================
// Helper: PostgreSQL Error Checks
// =============================================================================

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
