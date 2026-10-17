package presentation

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
)

type fakeHealthQueryService struct{ err error }

func (f fakeHealthQueryService) PingMongo(context.Context) error { return f.err }

func TestHealth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "mongo reachable", err: nil, want: http.StatusOK},
		{name: "mongo unreachable", err: errors.New("down"), want: http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := health.NewCheckUseCase(fakeHealthQueryService{err: tt.err}, health.StatusDTOFactory{}, logger)
			h := NewRouter(Controllers{Health: controller.NewHealthController(check)})
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
