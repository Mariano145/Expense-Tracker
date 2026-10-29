// Package service implements the business logic layer.
// AuthService orchestrates registration, login, token refresh, and logout
// using the repository interfaces for data access and the auth package for
// cryptographic operations.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Mariano145/spendly/backend/internal/auth"
	"github.com/Mariano145/spendly/backend/internal/model"
	"github.com/Mariano145/spendly/backend/internal/repository"
	"github.com/google/uuid"
)

// AuthService handles authentication business logic.
type AuthService struct {
	users       repository.UserRepository
	tokens      repository.RefreshTokenRepository
	jwtSecret   string
	accessExp   time.Duration
	refreshExp  time.Duration
	bcryptCost  int
}

// NewAuthService constructs an AuthService with its dependencies.
func NewAuthService(
	users repository.UserRepository,
	tokens repository.RefreshTokenRepository,
	jwtSecret string,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
	bcryptCost int,
) *AuthService {
	return &AuthService{
		users:      users,
		tokens:     tokens,
		jwtSecret:  jwtSecret,
		accessExp:  accessExpiry,
		refreshExp: refreshExpiry,
		bcryptCost: bcryptCost,
	}
}

// Register hashes the password and persists a new user.
func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	hash, err := auth.HashPassword(password, s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Email:        email,
		PasswordHash: hash,
	}
	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, model.ErrEmailExists) {
			return nil, model.ErrEmailExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// Login verifies credentials and returns an access token + refresh token pair.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return "", "", model.ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("find user: %w", err)
	}

	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		return "", "", model.ErrInvalidCredentials
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, s.jwtSecret, s.accessExp)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.createRefreshToken(ctx, user.ID, "")
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Refresh validates the presented refresh token and rotates it:
//   - If the token is already revoked → reuse detected → revoke entire family.
//   - If the token is expired → reject.
//   - Otherwise → revoke old token, issue new pair within the same family.
func (s *AuthService) Refresh(ctx context.Context, plainToken string) (string, string, error) {
	hash := auth.HashRefreshToken(plainToken)

	stored, err := s.tokens.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, model.ErrTokenNotFound) {
			return "", "", model.ErrTokenNotFound
		}
		return "", "", fmt.Errorf("find token: %w", err)
	}

	// Reuse detection: if the token is already revoked, someone is replaying it.
	if stored.Revoked {
		if revokeErr := s.tokens.RevokeByFamilyID(ctx, stored.FamilyID); revokeErr != nil {
			return "", "", fmt.Errorf("revoke family on reuse: %w", revokeErr)
		}
		return "", "", model.ErrTokenReuse
	}

	if stored.ExpiresAt.Before(time.Now()) {
		return "", "", model.ErrTokenExpired
	}

	// Revoke the old token.
	if err := s.tokens.RevokeByHash(ctx, hash); err != nil {
		return "", "", fmt.Errorf("revoke old token: %w", err)
	}

	// Look up user for the access token claims.
	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return "", "", fmt.Errorf("find user for refresh: %w", err)
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, s.jwtSecret, s.accessExp)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	newRefresh, err := s.createRefreshToken(ctx, user.ID, stored.FamilyID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefresh, nil
}

// Logout revokes a refresh token by its SHA-256 hash.
func (s *AuthService) Logout(ctx context.Context, plainToken string) error {
	hash := auth.HashRefreshToken(plainToken)
	if err := s.tokens.RevokeByHash(ctx, hash); err != nil {
		if errors.Is(err, model.ErrTokenNotFound) {
			return model.ErrTokenNotFound
		}
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

// ValidateToken parses and validates a JWT access token, returning the subject and email.
func (s *AuthService) ValidateToken(_ context.Context, tokenString string) (string, string, error) {
	return auth.ParseAccessToken(tokenString, s.jwtSecret)
}

// createRefreshToken generates a random refresh token, hashes it, and persists it.
// If familyID is empty, a new family is created (login). Otherwise the existing
// family is continued (rotation).
func (s *AuthService) createRefreshToken(ctx context.Context, userID, familyID string) (string, error) {
	plain, err := auth.GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	if familyID == "" {
		familyID = uuid.New().String()
	}

	rt := &model.RefreshToken{
		UserID:    userID,
		TokenHash: auth.HashRefreshToken(plain),
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(s.refreshExp),
	}
	if err := s.tokens.Create(ctx, rt); err != nil {
		return "", fmt.Errorf("persist refresh token: %w", err)
	}

	return plain, nil
}
