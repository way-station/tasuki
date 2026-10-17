package health

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

type fakeQueryService struct{ err error }

func (f fakeQueryService) PingMongo(context.Context) error { return f.err }

func TestCheckUseCase(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "mongo reachable", err: nil, want: true},
		{name: "mongo unreachable", err: errors.New("down"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewCheckUseCase(fakeQueryService{err: tt.err}, StatusDTOFactory{}, logger)
			got := uc.Execute(context.Background())
			if got.Healthy() != tt.want || got.MongoReachable() != tt.want {
				t.Errorf("Execute() healthy = %v, mongo = %v, want %v", got.Healthy(), got.MongoReachable(), tt.want)
			}
		})
	}
}
