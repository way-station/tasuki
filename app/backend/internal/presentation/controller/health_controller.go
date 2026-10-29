// Package controller holds the HTTP controllers. Each controller calls a use case
// and writes its result as the response.
package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/health"
)

// HealthController serves GET /health.
type HealthController struct {
	check *health.CheckUseCase
}

// NewHealthController returns a HealthController that runs check.
func NewHealthController(check *health.CheckUseCase) *HealthController {
	return &HealthController{check: check}
}

// Get writes 200 when every store is reachable, and 503 otherwise.
func (controller *HealthController) Get(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	status := controller.check.Execute(ctx)
	body := map[string]string{
		"status":   "ok",
		"mongo":    storeState(status.MongoReachable()),
		"postgres": storeState(status.PostgresReachable()),
	}
	if !status.Healthy() {
		body["status"] = "error"
		writeJSON(w, http.StatusServiceUnavailable, body)
		return
	}
	writeJSON(w, http.StatusOK, body)
}

func storeState(reachable bool) string {
	if reachable {
		return "ok"
	}
	return "unreachable"
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
