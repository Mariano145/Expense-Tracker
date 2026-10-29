package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Mariano145/spendly/backend/internal/auth"
	"github.com/Mariano145/spendly/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSecret is used across service tests for JWT signing.
const testSecret = "test-secret-that-is-at-least-32-chars-long"

// newTestService creates an AuthService with mock repositories for unit testing.
func newTestService(userRepo *mockUserRepo, tokenRepo *mockRefreshTokenRepo) *AuthService {
	return NewAuthService(userRepo, tokenRepo, testSecret, 15*time.Minute, 168*time.Hour, 10)
}

// --- Register ---

func TestAuthService_Register(t *testing.T) {
	t.Run("success - creates user with hashed password", func(t *testing.T) {
		repo := &mockUserRepo{
			createFn: func(ctx context.Context, user *model.User) error {
				user.ID = "user-1"
				user.CreatedAt = time.Now()
				user.UpdatedAt = time.Now()
				return nil
			},
		}
		svc := newTestService(repo, nil)

		user, err := svc.Register(context.Background(), "new@example.com", "strongP@ss1")
		require.NoError(t, err)
		assert.Equal(t, "user-1", user.ID)
		assert.Equal(t, "new@example.com", user.Email)
		assert.NotEmpty(t, user.PasswordHash, "password must be hashed before storage")
		assert.NotEqual(t, "strongP@ss1", user.PasswordHash, "plain-text password must not be stored")
	})

	t.Run("duplicate email - returns ErrEmailExists", func(t *testing.T) {
		repo := &mockUserRepo{
			createFn: func(ctx context.Context, user *model.User) error {
				return fmt.Errorf("create user: %w", model.ErrEmailExists)
			},
		}
		svc := newTestService(repo, nil)

		_, err := svc.Register(context.Background(), "taken@example.com", "password123")
		require.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrEmailExists), "expected ErrEmailExists, got: %v", err)
	})
}

// --- Login ---

func TestAuthService_Login(t *testing.T) {
	t.Run("success - returns access and refresh tokens", func(t *testing.T) {
		hash, _ := auth.HashPassword("correctPass", 10)
		userRepo := &mockUserRepo{
			findByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
				return &model.User{ID: "user-1", Email: email, PasswordHash: hash}, nil
			},
		}
		tokenRepo := &mockRefreshTokenRepo{
			createFn: func(ctx context.Context, token *model.RefreshToken) error {
				token.ID = "rt-1"
				token.CreatedAt = time.Now()
				return nil
			},
		}
		svc := newTestService(userRepo, tokenRepo)

		access, refresh, err := svc.Login(context.Background(), "user@example.com", "correctPass")
		require.NoError(t, err)
		assert.NotEmpty(t, access, "access token must be returned")
		assert.NotEmpty(t, refresh, "refresh token must be returned")

		// Verify the access token is parseable.
		userID, email, err := auth.ParseAccessToken(access, testSecret)
		require.NoError(t, err)
		assert.Equal(t, "user-1", userID)
		assert.Equal(t, "user@example.com", email)
	})

	t.Run("user not found - returns ErrInvalidCredentials", func(t *testing.T) {
		userRepo := &mockUserRepo{
			findByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
				return nil, model.ErrUserNotFound
			},
		}
		svc := newTestService(userRepo, nil)

		_, _, err := svc.Login(context.Background(), "ghost@example.com", "password")
		require.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrInvalidCredentials), "expected ErrInvalidCredentials, got: %v", err)
	})

	t.Run("wrong password - returns ErrInvalidCredentials", func(t *testing.T) {
		hash, _ := auth.HashPassword("correctPass", 10)
		userRepo := &mockUserRepo{
			findByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
				return &model.User{ID: "user-1", Email: email, PasswordHash: hash}, nil
			},
		}
		svc := newTestService(userRepo, nil)

		_, _, err := svc.Login(context.Background(), "user@example.com", "wrongPass")
		require.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrInvalidCredentials), "expected ErrInvalidCredentials, got: %v", err)
	})
}

// --- Refresh ---

func TestAuthService_Refresh(t *testing.T) {
	t.Run("success - rotates token within same family", func(t *testing.T) {
		oldHash := auth.HashRefreshToken("old-refresh-token")
		storedToken := &model.RefreshToken{
			ID:        "rt-old",
			UserID:    "user-1",
			TokenHash: oldHash,
			FamilyID:  "family-abc",
			ExpiresAt: time.Now().Add(100 * time.Hour),
			Revoked:   false,
		}

		userRepo := &mockUserRepo{
			findByIDFn: func(ctx context.Context, id string) (*model.User, error) {
				return &model.User{ID: id, Email: "user@example.com"}, nil
			},
		}
		tokenRepo := &mockRefreshTokenRepo{
			findByHashFn: func(ctx context.Context, hash string) (*model.RefreshToken, error) {
				return storedToken, nil
			},
			revokeByHashFn: func(ctx context.Context, hash string) error {
				storedToken.Revoked = true
				return nil
			},
			createFn: func(ctx context.Context, token *model.RefreshToken) error {
				token.ID = "rt-new"
				token.CreatedAt = time.Now()
				return nil
			},
		}
		svc := newTestService(userRepo, tokenRepo)

		newAccess, newRefresh, err := svc.Refresh(context.Background(), "old-refresh-token")
		require.NoError(t, err)
		assert.NotEmpty(t, newAccess)
		assert.NotEmpty(t, newRefresh)
		assert.True(t, storedToken.Revoked, "old token must be revoked after rotation")
	})

	t.Run("expired token - returns ErrTokenExpired", func(t *testing.T) {
		hash := auth.HashRefreshToken("expired-token")
		storedToken := &model.RefreshToken{
			ID:        "rt-exp",
			UserID:    "user-1",
			TokenHash: hash,
			FamilyID:  "family-xyz",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // already expired
			Revoked:   false,
		}
		tokenRepo := &mockRefreshTokenRepo{
			findByHashFn: func(ctx context.Context, h string) (*model.RefreshToken, error) {
				return storedToken, nil
			},
		}
		svc := newTestService(nil, tokenRepo)

		_, _, err := svc.Refresh(context.Background(), "expired-token")
		require.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrTokenExpired), "expected ErrTokenExpired, got: %v", err)
	})

	t.Run("revoked token (reuse detected) - revokes family and returns ErrTokenReuse", func(t *testing.T) {
		hash := auth.HashRefreshToken("stolen-token")
		var familyRevoked bool
		storedToken := &model.RefreshToken{
			ID:        "rt-stolen",
			UserID:    "user-1",
			TokenHash: hash,
			FamilyID:  "family-theft",
			ExpiresAt: time.Now().Add(100 * time.Hour),
			Revoked:   true, // already revoked — this is reuse!
		}
		tokenRepo := &mockRefreshTokenRepo{
			findByHashFn: func(ctx context.Context, h string) (*model.RefreshToken, error) {
				return storedToken, nil
			},
			revokeByFamilyFn: func(ctx context.Context, familyID string) error {
				familyRevoked = true
				return nil
			},
		}
		svc := newTestService(nil, tokenRepo)

		_, _, err := svc.Refresh(context.Background(), "stolen-token")
		require.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrTokenReuse), "expected ErrTokenReuse, got: %v", err)
		assert.True(t, familyRevoked, "entire family must be revoked on reuse detection")
	})

	t.Run("token not found - returns ErrTokenNotFound", func(t *testing.T) {
		tokenRepo := &mockRefreshTokenRepo{
			findByHashFn: func(ctx context.Context, h string) (*model.RefreshToken, error) {
				return nil, model.ErrTokenNotFound
			},
		}
		svc := newTestService(nil, tokenRepo)

		_, _, err := svc.Refresh(context.Background(), "nonexistent-token")
		require.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrTokenNotFound), "expected ErrTokenNotFound, got: %v", err)
	})
}

// --- Logout ---

func TestAuthService_Logout(t *testing.T) {
	t.Run("success - revokes the refresh token", func(t *testing.T) {
		var revokedHash string
		tokenRepo := &mockRefreshTokenRepo{
			revokeByHashFn: func(ctx context.Context, hash string) error {
				revokedHash = hash
				return nil
			},
		}
		svc := newTestService(nil, tokenRepo)

		err := svc.Logout(context.Background(), "some-refresh-token")
		require.NoError(t, err)
		expectedHash := auth.HashRefreshToken("some-refresh-token")
		assert.Equal(t, expectedHash, revokedHash, "must revoke by the token's SHA-256 hash")
	})

	t.Run("token not found - returns ErrTokenNotFound", func(t *testing.T) {
		tokenRepo := &mockRefreshTokenRepo{
			revokeByHashFn: func(ctx context.Context, hash string) error {
				return model.ErrTokenNotFound
			},
		}
		svc := newTestService(nil, tokenRepo)

		err := svc.Logout(context.Background(), "unknown-token")
		require.Error(t, err)
		assert.True(t, errors.Is(err, model.ErrTokenNotFound), "expected ErrTokenNotFound, got: %v", err)
	})
}

// --- ValidateToken ---

func TestAuthService_ValidateToken(t *testing.T) {
	t.Run("valid token - returns userID and email", func(t *testing.T) {
		token, err := auth.GenerateAccessToken("user-42", "valid@example.com", testSecret, 15*time.Minute)
		require.NoError(t, err)

		svc := newTestService(nil, nil)
		userID, email, err := svc.ValidateToken(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, "user-42", userID)
		assert.Equal(t, "valid@example.com", email)
	})

	t.Run("expired token - returns error", func(t *testing.T) {
		token, err := auth.GenerateAccessToken("user-42", "exp@example.com", testSecret, -5*time.Minute)
		require.NoError(t, err)

		svc := newTestService(nil, nil)
		_, _, err = svc.ValidateToken(context.Background(), token)
		require.Error(t, err)
	})

	t.Run("invalid token - returns error", func(t *testing.T) {
		svc := newTestService(nil, nil)
		_, _, err := svc.ValidateToken(context.Background(), "garbage-token")
		require.Error(t, err)
	})
}
