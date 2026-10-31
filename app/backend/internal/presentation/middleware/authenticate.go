// Package middleware holds the HTTP middleware of the backend.
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
)

// SessionCookieName is the cookie that carries the session token (ADR-0014).
const SessionCookieName = "tasuki_session"

// Authenticator identifies the user of a session token.
type Authenticator interface {
	Execute(ctx context.Context, token string) (shared.UserID, error)
}

type contextKey int

const (
	userIDKey contextKey = iota
	sessionTokenKey
)

// Authenticate reads the session cookie, identifies the user through authenticator,
// and passes it to the handlers in the request context. A missing, unknown, expired, or
// revoked session is 401 unauthenticated (ADR-0014). It does not read the database itself.
func Authenticate(authenticator Authenticator, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var token string
			if cookie, err := req.Cookie(SessionCookieName); err == nil {
				token = cookie.Value
			}
			userID, err := authenticator.Execute(req.Context(), token)
			if errors.Is(err, auth.ErrUnauthenticated) {
				WriteUnauthenticated(w)
				return
			}
			if err != nil {
				logger.ErrorContext(req.Context(), "authenticate failed", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(req.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, sessionTokenKey, token)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// UserID returns the user Authenticate identified.
func UserID(ctx context.Context) (shared.UserID, bool) {
	userID, ok := ctx.Value(userIDKey).(shared.UserID)
	return userID, ok
}

// SessionToken returns the session token of the request Authenticate accepted.
func SessionToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(sessionTokenKey).(string)
	return token, ok
}

// UnauthenticatedProblem is the body of a 401 response (docs/design/api.md).
func UnauthenticatedProblem() api.Problem {
	return api.Problem{Type: "about:blank", Title: "Unauthenticated", Status: http.StatusUnauthorized, Code: api.ProblemCodeUnauthenticated}
}

// WriteUnauthenticated writes a 401 problem+json response.
func WriteUnauthenticated(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(UnauthenticatedProblem())
}
