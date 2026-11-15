package middleware

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fakeCallerResolver struct {
	caller auth.MeDTO
	err    error
}

func (fake fakeCallerResolver) Execute(context.Context, shared.UserID) (auth.MeDTO, error) {
	return fake.caller, fake.err
}

func callerWith(isOrganizer bool, status identity.CallerStatus) auth.MeDTO {
	factory := auth.MeDTOFactory{}
	return factory.Create("u1", "Taro", isOrganizer, status, factory.CreateRosterMember("m1", "Taro"))
}

// serve runs handler for a request of userID, or of no user when userID is empty.
func serve(handler http.Handler, userID shared.UserID) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/races", nil)
	if userID != "" {
		req = req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestResolveCaller(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var got auth.MeDTO
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		got, _ = Caller(req.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	t.Run("passes the caller on", func(t *testing.T) {
		caller := callerWith(false, identity.CallerApproved)
		recorder := serve(ResolveCaller(fakeCallerResolver{caller: caller}, logger)(next), "u1")
		if recorder.Code != http.StatusNoContent || got.UserID() != "u1" || got.Status() != identity.CallerApproved {
			t.Errorf("status %d, caller %+v; want 204 and the resolved caller", recorder.Code, got)
		}
	})

	tests := []struct {
		name   string
		userID shared.UserID
		err    error
		want   int
	}{
		{name: "no user", userID: "", want: http.StatusUnauthorized},
		{name: "a user that no longer exists", userID: "u1", err: auth.ErrUnauthenticated, want: http.StatusUnauthorized},
		{name: "a failure", userID: "u1", err: errors.New("database is down"), want: http.StatusInternalServerError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := serve(ResolveCaller(fakeCallerResolver{err: testCase.err}, logger)(next), testCase.userID)
			if recorder.Code != testCase.want {
				t.Errorf("status = %d, want %d", recorder.Code, testCase.want)
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name          string
		caller        auth.MeDTO
		wantMember    int
		wantOrganizer int
	}{
		{name: "member", caller: callerWith(false, identity.CallerApproved), wantMember: http.StatusNoContent, wantOrganizer: http.StatusForbidden},
		{name: "organizer without a link", caller: callerWith(true, identity.CallerNone), wantMember: http.StatusNoContent, wantOrganizer: http.StatusNoContent},
		{name: "user waiting for approval", caller: callerWith(false, identity.CallerPending), wantMember: http.StatusForbidden, wantOrganizer: http.StatusForbidden},
		{name: "rejected user", caller: callerWith(false, identity.CallerRejected), wantMember: http.StatusForbidden, wantOrganizer: http.StatusForbidden},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			resolve := ResolveCaller(fakeCallerResolver{caller: testCase.caller}, logger)
			if code := serve(resolve(RequireMember(next)), "u1").Code; code != testCase.wantMember {
				t.Errorf("RequireMember status = %d, want %d", code, testCase.wantMember)
			}
			if code := serve(resolve(RequireOrganizer(next)), "u1").Code; code != testCase.wantOrganizer {
				t.Errorf("RequireOrganizer status = %d, want %d", code, testCase.wantOrganizer)
			}
		})
	}

	t.Run("refuses a request without a resolved caller", func(t *testing.T) {
		recorder := serve(RequireMember(next), "u1")
		if recorder.Code != http.StatusForbidden || recorder.Header().Get("Content-Type") != "application/problem+json" {
			t.Errorf("status %d, Content-Type %q; want a 403 problem", recorder.Code, recorder.Header().Get("Content-Type"))
		}
	})
}
