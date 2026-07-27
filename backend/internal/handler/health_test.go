package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantBody   HealthResponse
	}{
		{
			name:       "GET /health returns 200 with ok status",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantBody:   HealthResponse{Status: "ok"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/health", nil)
			rec := httptest.NewRecorder()

			Health(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var got HealthResponse
			err := json.NewDecoder(rec.Body).Decode(&got)
			require.NoError(t, err)
			assert.Equal(t, tc.wantBody, got)
		})
	}
}
