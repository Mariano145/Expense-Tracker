// Package handler provides HTTP handlers for the billetera API.
package handler

import (
	"encoding/json"
	"net/http"
)

// HealthResponse is the JSON body returned by the health check endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// Health returns a 200 OK with {"status":"ok"} to signal the service is running.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}
