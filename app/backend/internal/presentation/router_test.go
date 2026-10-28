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

func (f fakeMongoQueryService) PingMongo(context.Context) error { return f.err }

type fakePostgresQueryService struct{ err error }

func (f fakePostgresQueryService) PingPostgres(context.Context) error { return f.err }

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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := health.NewCheckUseCase(fakeMongoQueryService{err: tt.mongoErr}, fakePostgresQueryService{err: tt.postgresErr}, health.StatusDTOFactory{}, logger)
			h := NewRouter(Controllers{Health: controller.NewHealthController(check)})
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
			if got := strings.TrimSpace(rec.Body.String()); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}
