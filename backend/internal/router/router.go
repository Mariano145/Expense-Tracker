// Package router sets up HTTP routing for the billetera API.
// It uses the standard library's http.ServeMux with Go 1.22+ method patterns.
package router

import (
	"net/http"

	"github.com/gentle-ai/billetera/backend/internal/handler"
)

// New creates and returns a configured http.Handler with all API routes.
func New() http.Handler {
	mux := http.NewServeMux()

	// Health check — used by Docker healthcheck and CI smoke tests.
	mux.HandleFunc("GET /health", handler.Health)

	return mux
}
