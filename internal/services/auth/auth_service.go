package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/arilbois/contentbank-v2/internal/models"
	"github.com/arilbois/contentbank-v2/internal/repositories"
)

// Common errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

// Claims is the JWT payload.
type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"usr"`
	Role     string `json:"rol"`
	jwt.RegisteredClaims
}

// Service handles authentication logic.
type Service struct {
	users    *repositories.UserRepository
	secret   []byte
	tokenTTL time.Duration
}

func NewService(users *repositories.UserRepository, secret string, ttl time.Duration) *Service {
	return &Service{
		users:    users,
		secret:   []byte(secret),
		tokenTTL: ttl,
	}
}

// HashPassword returns a bcrypt hash of the plaintext password.
func (s *Service) HashPassword(plain string) (string, error) {
	if len(plain) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters")
	}
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

// CheckPassword returns nil if the plaintext matches the stored hash.
func (s *Service) CheckPassword(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return fmt.Errorf("compare: %w", ErrInvalidCredentials)
	}
	return nil
}

// Register creates a new user. Returns ErrUserExists if the username is taken.
func (s *Service) Register(ctx context.Context, username, password, role string) (*models.User, error) {
	if existing, err := s.users.GetByUsername(ctx, username); err == nil && existing != nil {
		return nil, ErrUserExists
	} else if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}

	hash, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}
	if role == "" {
		role = "viewer"
	}
	u := &models.User{
		Username:     username,
		PasswordHash: hash,
		Role:         role,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// Login authenticates a user and returns the user + a signed JWT.
func (s *Service) Login(ctx context.Context, username, password string) (*models.User, string, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("get user: %w", err)
	}
	if err := s.CheckPassword(u.PasswordHash, password); err != nil {
		return nil, "", ErrInvalidCredentials
	}
	tok, err := s.GenerateToken(u)
	if err != nil {
		return nil, "", err
	}
	return u, tok, nil
}

// GenerateToken returns a signed JWT for the given user.
func (s *Service) GenerateToken(u *models.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   u.ID.String(),
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			Issuer:    "contentbank-v2",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ValidateToken parses and validates a JWT, returning the claims.
func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
