package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
)

// RequireOrigin protects against CSRF (ADR-0014): a request that changes data (POST, PUT,
// PATCH, DELETE) must carry an Origin header equal to frontendOrigin, otherwise it is
// 403 forbidden. Browsers always send Origin on these requests, so a missing one is
// refused too. GET, HEAD, and OPTIONS change nothing and pass.
func RequireOrigin(frontendOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if changesData(req.Method) && req.Header.Get("Origin") != frontendOrigin {
				writeForbiddenOrigin(w)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

func changesData(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func writeForbiddenOrigin(w http.ResponseWriter) {
	detail := "a request that changes data must come from the frontend's origin"
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(api.Problem{Type: "about:blank", Title: "Forbidden", Status: http.StatusForbidden, Code: api.ProblemCodeForbidden, Detail: &detail})
}
