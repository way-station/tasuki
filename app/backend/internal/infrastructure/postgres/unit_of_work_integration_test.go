//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

func TestUnitOfWork(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	ctx := context.Background()
	factory := race.NewFactory(random.IDGenerator{})
	repository := NewRaceRepository(client, factory)
	unitOfWork := NewUnitOfWork(client)

	newRace := func(t *testing.T) *race.Race {
		t.Helper()
		created, err := factory.New(testTeamID, "Ekiden", distanceDescription(t, 5))
		if err != nil {
			t.Fatal(err)
		}
		return created
	}

	t.Run("commits when the work succeeds, including nested work", func(t *testing.T) {
		outer, inner := newRace(t), newRace(t)
		err := unitOfWork.Run(ctx, func(ctx context.Context) error {
			if err := repository.Save(ctx, outer); err != nil {
				return err
			}
			return unitOfWork.Run(ctx, func(ctx context.Context) error { return repository.Save(ctx, inner) })
		})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		for _, saved := range []*race.Race{outer, inner} {
			if _, err := repository.FindByID(ctx, saved.ID()); err != nil {
				t.Errorf("FindByID after commit: %v", err)
			}
		}
	})

	t.Run("rolls back every write when the work fails", func(t *testing.T) {
		saved := newRace(t)
		workErr := errors.New("adjusting the running orders failed")
		err := unitOfWork.Run(ctx, func(ctx context.Context) error {
			if err := repository.Save(ctx, saved); err != nil {
				return err
			}
			if _, err := repository.FindByID(ctx, saved.ID()); err != nil {
				t.Errorf("FindByID inside the transaction: %v", err)
			}
			return workErr
		})
		if !errors.Is(err, workErr) {
			t.Errorf("Run error = %v, want %v", err, workErr)
		}
		if _, err := repository.FindByID(ctx, saved.ID()); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("FindByID after rollback error = %v, want ErrNotFound", err)
		}
	})
}
