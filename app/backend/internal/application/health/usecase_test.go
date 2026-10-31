package health

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

type fakeMongoQueryService struct{ err error }

func (fake fakeMongoQueryService) PingMongo(context.Context) error { return fake.err }

type fakePostgresQueryService struct{ err error }

func (fake fakePostgresQueryService) PingPostgres(context.Context) error { return fake.err }

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
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			useCase := NewCheckUseCase(fakeMongoQueryService{err: testCase.mongoErr}, fakePostgresQueryService{err: testCase.postgresErr}, StatusDTOFactory{}, logger)
			got := useCase.Execute(context.Background())
			if got.Healthy() != testCase.wantHealthy || got.MongoReachable() != testCase.wantMongo || got.PostgresReachable() != testCase.wantPostgres {
				t.Errorf("Execute() healthy = %v, mongo = %v, postgres = %v, want %v, %v, %v",
					got.Healthy(), got.MongoReachable(), got.PostgresReachable(), testCase.wantHealthy, testCase.wantMongo, testCase.wantPostgres)
			}
		})
	}
}
