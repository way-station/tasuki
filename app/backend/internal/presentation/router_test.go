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
			handler := NewRouter(Controllers{Health: controller.NewHealthController(check)})
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
