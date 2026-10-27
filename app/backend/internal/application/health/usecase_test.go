package health

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

type fakeMongoQueryService struct{ err error }

func (f fakeMongoQueryService) PingMongo(context.Context) error { return f.err }

type fakePostgresQueryService struct{ err error }

func (f fakePostgresQueryService) PingPostgres(context.Context) error { return f.err }

func TestCheckUseCase(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	down := errors.New("down")

	tests := []struct {
		name         string
		mongoErr     error
		postgresErr  error
		wantHealthy  bool
		wantMongo    bool
		wantPostgres bool
	}{
		{name: "both reachable", wantHealthy: true, wantMongo: true, wantPostgres: true},
		{name: "mongo unreachable", mongoErr: down, wantPostgres: true},
		{name: "postgres unreachable", postgresErr: down, wantMongo: true},
		{name: "both unreachable", mongoErr: down, postgresErr: down},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewCheckUseCase(fakeMongoQueryService{err: tt.mongoErr}, fakePostgresQueryService{err: tt.postgresErr}, StatusDTOFactory{}, logger)
			got := uc.Execute(context.Background())
			if got.Healthy() != tt.wantHealthy || got.MongoReachable() != tt.wantMongo || got.PostgresReachable() != tt.wantPostgres {
				t.Errorf("Execute() healthy = %v, mongo = %v, postgres = %v, want %v, %v, %v",
					got.Healthy(), got.MongoReachable(), got.PostgresReachable(), tt.wantHealthy, tt.wantMongo, tt.wantPostgres)
			}
		})
	}
}
