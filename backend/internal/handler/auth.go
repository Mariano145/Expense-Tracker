package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Mariano145/spendly/backend/internal/middleware"
	"github.com/Mariano145/spendly/backend/internal/model"
)

type authService interface {
	Register(ctx context.Context, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(ctx context.Context, tokenString string) (string, string, error)
}

type AuthHandler struct {
	svc authService
}

func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if msg := validateRegistration(req.Email, req.Password); msg != "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", msg)
		return
	}

	user, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toUserDTO(user))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	if msg := validateLogin(req.Email, req.Password); msg != "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", msg)
		return
	}

	accessToken, refreshToken, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	setRefreshCookie(w, refreshToken)
	writeJSON(w, http.StatusOK, AuthResponse{
		AccessToken: accessToken,
		User:        UserDTO{Email: req.Email},
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing refresh token")
		return
	}

	newAccess, newRefresh, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	setRefreshCookie(w, newRefresh)
	writeJSON(w, http.StatusOK, RefreshResponse{AccessToken: newAccess})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing refresh token")
		return
	}

	if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
		if errors.Is(err, model.ErrTokenNotFound) {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "token not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not logout")
		return
	}

	clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		return
	}
	email, _ := r.Context().Value(middleware.ContextKeyEmail).(string)

	writeJSON(w, http.StatusOK, UserDTO{ID: userID, Email: email})
}

func setRefreshCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
}

func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func toUserDTO(u *model.User) UserDTO {
	return UserDTO{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrEmailExists):
		writeError(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, model.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, model.ErrTokenExpired):
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, model.ErrTokenRevoked):
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, model.ErrTokenNotFound):
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	case errors.Is(err, model.ErrTokenReuse):
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{
		Error: ErrorDetail{Code: code, Message: message},
	})
}
