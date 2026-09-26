package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTokenValidator struct {
	validateFn func(ctx context.Context, token string) (string, string, error)
}

func (m *mockTokenValidator) ValidateToken(ctx context.Context, token string) (string, string, error) {
	return m.validateFn(ctx, token)
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("valid token injects userID and email into context", func(t *testing.T) {
		validator := &mockTokenValidator{
			validateFn: func(_ context.Context, _ string) (string, string, error) {
				return "user-123", "test@test.com", nil
			},
		}

		var capturedUserID, capturedEmail string
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUserID, _ = r.Context().Value(ContextKeyUserID).(string)
			capturedEmail, _ = r.Context().Value(ContextKeyEmail).(string)
			w.WriteHeader(http.StatusOK)
		})

		handler := Auth(validator)(inner)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer valid-jwt-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "user-123", capturedUserID)
		assert.Equal(t, "test@test.com", capturedEmail)
	})

	t.Run("missing Authorization header returns 401", func(t *testing.T) {
		validator := &mockTokenValidator{}
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("inner handler should not be called")
		})

		handler := Auth(validator)(inner)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("malformed Authorization header returns 401", func(t *testing.T) {
		validator := &mockTokenValidator{}
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("inner handler should not be called")
		})

		handler := Auth(validator)(inner)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Authorization", "NotBearer token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		validator := &mockTokenValidator{
			validateFn: func(_ context.Context, _ string) (string, string, error) {
				return "", "", assert.AnError
			},
		}
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("inner handler should not be called")
		})

		handler := Auth(validator)(inner)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestCORSMiddleware(t *testing.T) {
	t.Run("sets CORS headers with configured origin", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := CORS("http://localhost:3000")(inner)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("preflight OPTIONS returns 204 with headers", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("inner handler should not be called for preflight")
		})

		handler := CORS("http://localhost:3000")(inner)

		req := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
		require.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
		require.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		require.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("non-matching origin does not set Allow-Origin", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := CORS("http://localhost:3000")(inner)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.Header.Set("Origin", "http://evil.com")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}
