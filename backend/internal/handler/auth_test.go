package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gentle-ai/billetera/backend/internal/middleware"
	"github.com/gentle-ai/billetera/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	registerFn      func(ctx context.Context, email, password string) (*model.User, error)
	loginFn         func(ctx context.Context, email, password string) (string, string, error)
	refreshFn       func(ctx context.Context, refreshToken string) (string, string, error)
	logoutFn        func(ctx context.Context, refreshToken string) error
	validateTokenFn func(ctx context.Context, tokenString string) (string, string, error)
}

func (m *mockAuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	return m.registerFn(ctx, email, password)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	return m.loginFn(ctx, email, password)
}

func (m *mockAuthService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	return m.refreshFn(ctx, refreshToken)
}

func (m *mockAuthService) Logout(ctx context.Context, refreshToken string) error {
	return m.logoutFn(ctx, refreshToken)
}

func (m *mockAuthService) ValidateToken(ctx context.Context, tokenString string) (string, string, error) {
	return m.validateTokenFn(ctx, tokenString)
}

func decodeError(t *testing.T, body *bytes.Buffer) ErrorResponse {
	t.Helper()
	var errResp ErrorResponse
	require.NoError(t, json.NewDecoder(body).Decode(&errResp))
	return errResp
}

func TestRegister(t *testing.T) {
	t.Run("success returns 201 with user", func(t *testing.T) {
		now := time.Now()
		svc := &mockAuthService{
			registerFn: func(_ context.Context, _, _ string) (*model.User, error) {
				return &model.User{ID: "u1", Email: "test@test.com", CreatedAt: now}, nil
			},
		}
		h := NewAuthHandler(svc)

		body, _ := json.Marshal(RegisterRequest{Email: "test@test.com", Password: "secure123"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp UserDTO
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "u1", resp.ID)
		assert.Equal(t, "test@test.com", resp.Email)
	})

	t.Run("duplicate email returns 409", func(t *testing.T) {
		svc := &mockAuthService{
			registerFn: func(_ context.Context, _, _ string) (*model.User, error) {
				return nil, model.ErrEmailExists
			},
		}
		h := NewAuthHandler(svc)

		body, _ := json.Marshal(RegisterRequest{Email: "dup@test.com", Password: "secure123"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
		errResp := decodeError(t, rec.Body)
		assert.Equal(t, "CONFLICT", errResp.Error.Code)
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		svc := &mockAuthService{}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("{bad")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("internal error returns 500", func(t *testing.T) {
		svc := &mockAuthService{
			registerFn: func(_ context.Context, _, _ string) (*model.User, error) {
				return nil, errors.New("db connection lost")
			},
		}
		h := NewAuthHandler(svc)

		body, _ := json.Marshal(RegisterRequest{Email: "test@test.com", Password: "secure123"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Register(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestLogin(t *testing.T) {
	t.Run("success returns 200 with token and cookie", func(t *testing.T) {
		now := time.Now()
		svc := &mockAuthService{
			loginFn: func(_ context.Context, _, _ string) (string, string, error) {
				return "access-jwt", "refresh-plain", nil
			},
			registerFn: nil,
		}
		h := NewAuthHandler(svc)

		body, _ := json.Marshal(LoginRequest{Email: "test@test.com", Password: "secure123"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		_ = now
		h.Login(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp AuthResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "access-jwt", resp.AccessToken)

		cookies := rec.Result().Cookies()
		var refreshCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}
		require.NotNil(t, refreshCookie, "refresh_token cookie must be set")
		assert.True(t, refreshCookie.HttpOnly)
		assert.True(t, refreshCookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, refreshCookie.SameSite)
		assert.Equal(t, "/api/auth", refreshCookie.Path)
		assert.Equal(t, "refresh-plain", refreshCookie.Value)
	})

	t.Run("invalid credentials returns 401", func(t *testing.T) {
		svc := &mockAuthService{
			loginFn: func(_ context.Context, _, _ string) (string, string, error) {
				return "", "", model.ErrInvalidCredentials
			},
		}
		h := NewAuthHandler(svc)

		body, _ := json.Marshal(LoginRequest{Email: "test@test.com", Password: "wrong"})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		errResp := decodeError(t, rec.Body)
		assert.Equal(t, "UNAUTHORIZED", errResp.Error.Code)
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		svc := &mockAuthService{}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("not-json")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestRefresh(t *testing.T) {
	t.Run("success returns 200 with new access token and cookie", func(t *testing.T) {
		svc := &mockAuthService{
			refreshFn: func(_ context.Context, _ string) (string, string, error) {
				return "new-access", "new-refresh", nil
			},
		}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh"})
		rec := httptest.NewRecorder()

		h.Refresh(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp RefreshResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "new-access", resp.AccessToken)

		cookies := rec.Result().Cookies()
		var refreshCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				refreshCookie = c
			}
		}
		require.NotNil(t, refreshCookie)
		assert.Equal(t, "new-refresh", refreshCookie.Value)
	})

	t.Run("missing cookie returns 401", func(t *testing.T) {
		svc := &mockAuthService{}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		rec := httptest.NewRecorder()

		h.Refresh(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		svc := &mockAuthService{
			refreshFn: func(_ context.Context, _ string) (string, string, error) {
				return "", "", model.ErrTokenExpired
			},
		}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "expired-token"})
		rec := httptest.NewRecorder()

		h.Refresh(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("token reuse returns 401", func(t *testing.T) {
		svc := &mockAuthService{
			refreshFn: func(_ context.Context, _ string) (string, string, error) {
				return "", "", model.ErrTokenReuse
			},
		}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "reused-token"})
		rec := httptest.NewRecorder()

		h.Refresh(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestLogout(t *testing.T) {
	t.Run("success returns 204 and clears cookie", func(t *testing.T) {
		svc := &mockAuthService{
			logoutFn: func(_ context.Context, _ string) error {
				return nil
			},
		}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "some-token"})
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)

		cookies := rec.Result().Cookies()
		var clearCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "refresh_token" {
				clearCookie = c
			}
		}
		require.NotNil(t, clearCookie, "must clear refresh_token cookie")
		assert.True(t, clearCookie.MaxAge <= 0, "cookie MaxAge should be <= 0 to clear")
	})

	t.Run("missing cookie returns 401", func(t *testing.T) {
		svc := &mockAuthService{}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockAuthService{
			logoutFn: func(_ context.Context, _ string) error {
				return errors.New("db error")
			},
		}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "some-token"})
		rec := httptest.NewRecorder()

		h.Logout(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestMe(t *testing.T) {
	t.Run("returns user from context", func(t *testing.T) {
		svc := &mockAuthService{}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, "u1")
		ctx = context.WithValue(ctx, middleware.ContextKeyEmail, "test@test.com")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp UserDTO
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, "u1", resp.ID)
		assert.Equal(t, "test@test.com", resp.Email)
	})

	t.Run("missing context returns 401", func(t *testing.T) {
		svc := &mockAuthService{}
		h := NewAuthHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		rec := httptest.NewRecorder()

		h.Me(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
