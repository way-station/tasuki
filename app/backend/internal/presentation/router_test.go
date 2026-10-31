package presentation

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
)

type fakeMongoQueryService struct{ err error }

func (fake fakeMongoQueryService) PingMongo(context.Context) error { return fake.err }

type fakePostgresQueryService struct{ err error }

func (fake fakePostgresQueryService) PingPostgres(context.Context) error { return fake.err }

func TestHealth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	down := errors.New("down")

	tests := []struct {
		name        string
		mongoErr    error
		postgresErr error
		want        int
		wantBody    string
	}{
		{name: "both reachable", want: http.StatusOK, wantBody: `{"mongo":"ok","postgres":"ok","status":"ok"}`},
		{name: "mongo unreachable", mongoErr: down, want: http.StatusServiceUnavailable, wantBody: `{"mongo":"unreachable","postgres":"ok","status":"error"}`},
		{name: "postgres unreachable", postgresErr: down, want: http.StatusServiceUnavailable, wantBody: `{"mongo":"ok","postgres":"unreachable","status":"error"}`},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			check := health.NewCheckUseCase(fakeMongoQueryService{err: testCase.mongoErr}, fakePostgresQueryService{err: testCase.postgresErr}, health.StatusDTOFactory{}, logger)
			handler := NewRouter(Controllers{Health: controller.NewHealthController(check)}, testOptions())
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

			if rec.Code != testCase.want {
				t.Errorf("status = %d, want %d", rec.Code, testCase.want)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != testCase.wantBody {
				t.Errorf("body = %s, want %s", got, testCase.wantBody)
			}
		})
	}
}

func TestRouterWithFrontend(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	check := health.NewCheckUseCase(fakeMongoQueryService{}, fakePostgresQueryService{}, health.StatusDTOFactory{}, logger)
	frontend := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "frontend")
	})
	options := testOptions()
	options.Frontend = frontend
	handler := NewRouter(Controllers{Health: controller.NewHealthController(check)}, options)

	tests := []struct {
		name     string
		method   string
		path     string
		want     int
		wantBody string
	}{
		{name: "frontend route", method: http.MethodGet, path: "/races/123", want: http.StatusOK, wantBody: "frontend"},
		{name: "health", method: http.MethodGet, path: "/health", want: http.StatusOK, wantBody: `{"mongo":"ok","postgres":"ok","status":"ok"}`},
		{name: "unknown API path", method: http.MethodGet, path: "/api/v1/unknown", want: http.StatusNotFound},
		{name: "frontend answers HEAD", method: http.MethodHead, path: "/races/123", want: http.StatusOK},
		{name: "frontend serves GET and HEAD only", method: http.MethodPost, path: "/races/123", want: http.StatusMethodNotAllowed},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(testCase.method, testCase.path, nil))

			if rec.Code != testCase.want {
				t.Fatalf("status = %d, want %d", rec.Code, testCase.want)
			}
			if testCase.wantBody != "" && strings.TrimSpace(rec.Body.String()) != testCase.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), testCase.wantBody)
			}
		})
	}
}

func TestRouterWithoutFrontend(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	check := health.NewCheckUseCase(fakeMongoQueryService{}, fakePostgresQueryService{}, health.StatusDTOFactory{}, logger)
	handler := NewRouter(Controllers{Health: controller.NewHealthController(check)}, testOptions())

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/races/123", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// testOptions are router options for tests that do not reach the API routes.
func testOptions() Options {
	return Options{
		Authenticate:   func(next http.Handler) http.Handler { return next },
		FrontendOrigin: testOrigin,
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}
