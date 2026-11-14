package races

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type transactionKey struct{}

func inTransaction(ctx context.Context) bool {
	marked, _ := ctx.Value(transactionKey{}).(bool)
	return marked
}

type fakeUnitOfWork struct{ err error }

func (fake *fakeUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	fake.err = work(context.WithValue(ctx, transactionKey{}, true))
	return fake.err
}

type transactionalRaceRepository struct {
	fakeRaceRepository
	savedInTransaction bool
}

func (fake *transactionalRaceRepository) Save(ctx context.Context, saved *race.Race) error {
	fake.savedInTransaction = inTransaction(ctx)
	return fake.fakeRaceRepository.Save(ctx, saved)
}

type fakeAdjuster struct {
	called        bool
	inTransaction bool
	raceID        shared.RaceID
	description   race.Description
	err           error
}

func (fake *fakeAdjuster) AdjustRunningOrders(ctx context.Context, raceID shared.RaceID, description race.Description) error {
	fake.called, fake.inTransaction, fake.raceID, fake.description = true, inTransaction(ctx), raceID, description
	return fake.err
}

func storedRace(t *testing.T, team shared.TeamID) *race.Race {
	t.Helper()
	command, err := RaceCommandFactory{}.CreateRace("Ekiden", distanceInput())
	if err != nil {
		t.Fatal(err)
	}
	return race.NewFactory(fixedIDs{}).Reconstruct("r1", team, "Ekiden", command.description, 2)
}

func TestUpdateRaceUseCase(t *testing.T) {
	update := func(t *testing.T, input DescriptionInput, version int) UpdateRaceCommand {
		t.Helper()
		command, err := RaceCommandFactory{}.UpdateRace("r1", "Ekiden 2026", input, version)
		if err != nil {
			t.Fatal(err)
		}
		return command
	}
	moreLegs := distanceInput()
	moreLegs.LegDistancesKm = []float64{5, 3.2, 4}
	otherDistances := distanceInput()
	otherDistances.LegDistancesKm = []float64{6, 2}

	tests := []struct {
		name   string
		input  DescriptionInput
		adjust bool
	}{
		{name: "same legs", input: otherDistances, adjust: false},
		{name: "more legs", input: moreLegs, adjust: true},
		{name: "distance to time", input: timeInput(), adjust: true},
	}
	for _, testCase := range tests {
		t.Run("saves the edit in the unit of work: "+testCase.name, func(t *testing.T) {
			races := &transactionalRaceRepository{fakeRaceRepository: fakeRaceRepository{stored: storedRace(t, teamID)}}
			adjuster := &fakeAdjuster{}
			err := NewUpdateRaceUseCase(&fakeUnitOfWork{}, races, adjuster, teamID).Execute(context.Background(), update(t, testCase.input, 2))
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if races.saved == nil || races.saved.Name() != "Ekiden 2026" || races.saved.Version() != 3 || !races.savedInTransaction {
				t.Errorf("saved %+v in transaction %v, want version 3 in the transaction", races.saved, races.savedInTransaction)
			}
			if adjuster.called != testCase.adjust {
				t.Fatalf("running orders adjusted = %v, want %v", adjuster.called, testCase.adjust)
			}
			if testCase.adjust && (!adjuster.inTransaction || adjuster.raceID != "r1" || adjuster.description.Type() != races.saved.Description().Type() || adjuster.description.LegCount() != races.saved.Description().LegCount()) {
				t.Errorf("adjusted %s to %+v in transaction %v, want r1 to the new description in the transaction", adjuster.raceID, adjuster.description, adjuster.inTransaction)
			}
		})
	}

	t.Run("rolls back when the running orders cannot be adjusted", func(t *testing.T) {
		adjustErr := errors.New("database is down")
		unitOfWork := &fakeUnitOfWork{}
		races := &fakeRaceRepository{stored: storedRace(t, teamID)}
		err := NewUpdateRaceUseCase(unitOfWork, races, &fakeAdjuster{err: adjustErr}, teamID).Execute(context.Background(), update(t, moreLegs, 2))
		if !errors.Is(err, adjustErr) || !errors.Is(unitOfWork.err, adjustErr) {
			t.Errorf("Execute() error = %v, unit of work got %v, want %v", err, unitOfWork.err, adjustErr)
		}
	})

	failures := []struct {
		name    string
		stored  *race.Race
		version int
		want    error
	}{
		{name: "the race changed since it was read", stored: storedRace(t, teamID), version: 1, want: shared.ErrVersionConflict},
		{name: "the race does not exist", stored: nil, version: 2, want: shared.ErrNotFound},
		{name: "the race belongs to another team", stored: storedRace(t, "t2"), version: 2, want: shared.ErrNotFound},
	}
	for _, testCase := range failures {
		t.Run(testCase.name, func(t *testing.T) {
			races := &fakeRaceRepository{stored: testCase.stored}
			adjuster := &fakeAdjuster{}
			err := NewUpdateRaceUseCase(&fakeUnitOfWork{}, races, adjuster, teamID).Execute(context.Background(), update(t, moreLegs, testCase.version))
			if !errors.Is(err, testCase.want) {
				t.Errorf("Execute() error = %v, want %v", err, testCase.want)
			}
			if races.saved != nil || adjuster.called {
				t.Error("the race was saved or its running orders were adjusted")
			}
		})
	}
}
