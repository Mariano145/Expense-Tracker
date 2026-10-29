// Package router sets up HTTP routing for the spendly API.
// It uses the standard library's http.ServeMux with Go 1.22+ method patterns.
package router

import (
	"net/http"

	"github.com/Mariano145/spendly/backend/internal/handler"
)

// Deps holds the dependencies required to build the router.
type Deps struct {
	AuthHandler *handler.AuthHandler
	AuthMW      func(http.Handler) http.Handler
	CORSMW      func(http.Handler) http.Handler
}

// New creates and returns a configured http.Handler with all API routes.
func New(deps Deps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)

	mux.HandleFunc("POST /api/auth/register", deps.AuthHandler.Register)
	mux.HandleFunc("POST /api/auth/login", deps.AuthHandler.Login)
	mux.HandleFunc("POST /api/auth/refresh", deps.AuthHandler.Refresh)
	mux.HandleFunc("POST /api/auth/logout", deps.AuthHandler.Logout)

	mux.Handle("GET /api/auth/me", deps.AuthMW(http.HandlerFunc(deps.AuthHandler.Me)))

	return deps.CORSMW(mux)
}
