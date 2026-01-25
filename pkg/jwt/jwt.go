package jwt

import (
	"errors"
	"time"

	"go-template-clean-architecture/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
	TokenID   string    `json:"token_id"`
	jwt.RegisteredClaims
}

type JWTService struct {
	config config.JWTConfig
}

func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{config: cfg}
}

func (s *JWTService) GenerateAccessToken(userID uuid.UUID, email string) (string, string, error) {
	tokenID := uuid.New().String()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		TokenType: AccessToken,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", "", err
	}

	return signedToken, tokenID, nil
}

func (s *JWTService) GenerateRefreshToken(userID uuid.UUID, email string) (string, string, error) {
	tokenID := uuid.New().String()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		TokenType: RefreshToken,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", "", err
	}

	return signedToken, tokenID, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *JWTService) GetAccessExpiry() time.Duration {
	return s.config.AccessExpiry
}

func (s *JWTService) GetRefreshExpiry() time.Duration {
	return s.config.RefreshExpiry
}
