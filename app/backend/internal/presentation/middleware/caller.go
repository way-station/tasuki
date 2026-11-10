package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
)

// CallerResolver tells who a user is: organizer or not, the link status, and the roster
// member they act as. GetMeUseCase is one.
type CallerResolver interface {
	Execute(ctx context.Context, userID shared.UserID) (auth.MeDTO, error)
}

type callerKey struct{}

// ResolveCaller resolves the caller once per request and passes it to the handlers in the
// request context (docs/design/api.md). It runs after Authenticate; a user that no longer
// exists is 401 unauthenticated.
func ResolveCaller(resolver CallerResolver, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			userID, ok := UserID(req.Context())
			if !ok {
				WriteUnauthenticated(w)
				return
			}
			caller, err := resolver.Execute(req.Context(), userID)
			if errors.Is(err, auth.ErrUnauthenticated) {
				WriteUnauthenticated(w)
				return
			}
			if err != nil {
				logger.ErrorContext(req.Context(), "resolve caller failed", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), callerKey{}, caller)))
		})
	}
}

// Caller returns the caller ResolveCaller resolved.
func Caller(ctx context.Context) (auth.MeDTO, bool) {
	caller, ok := ctx.Value(callerKey{}).(auth.MeDTO)
	return caller, ok
}

// RequireMember lets through members and organizers ("Who can call" in
// docs/design/api.md). Anyone else, such as a user waiting for approval, is 403 forbidden.
func RequireMember(next http.Handler) http.Handler {
	return requireRole(next, func(caller auth.MeDTO) bool {
		return caller.IsOrganizer() || caller.Status() == identity.CallerApproved
	})
}

// RequireOrganizer lets through organizers only. Anyone else is 403 forbidden.
func RequireOrganizer(next http.Handler) http.Handler {
	return requireRole(next, auth.MeDTO.IsOrganizer)
}

func requireRole(next http.Handler, allowed func(auth.MeDTO) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		caller, ok := Caller(req.Context())
		if !ok || !allowed(caller) {
			writeForbiddenRole(w)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func writeForbiddenRole(w http.ResponseWriter) {
	detail := "the caller's role does not allow this endpoint"
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(api.Problem{Type: "about:blank", Title: "Forbidden", Status: http.StatusForbidden, Code: api.ProblemCodeForbidden, Detail: &detail})
}
