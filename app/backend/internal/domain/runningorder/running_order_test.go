package runningorder

import (
	"errors"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "o1" }

var members = []shared.MemberID{"a", "b", "c"}

func runner(id shared.MemberID) *shared.MemberID { return &id }

func assign(runners ...*shared.MemberID) []Assignment {
	assignments := make([]Assignment, len(runners))
	for i, runnerID := range runners {
		assignments[i] = Assignment{Number: i + 1, Runner: runnerID}
	}
	return assignments
}

func runners(runningOrder *RunningOrder) []string {
	var got []string
	for _, leg := range runningOrder.Legs() {
		runnerID, ok := leg.Runner()
		if !ok {
			runnerID = "-"
		}
		got = append(got, string(runnerID))
	}
	return got
}

func isValidation(err error) bool {
	var validationErr *shared.ValidationError
	return errors.As(err, &validationErr)
}

func newConfirmed(t *testing.T, shape Shape, assignments []Assignment) *RunningOrder {
	t.Helper()
	runningOrder, err := NewFactory(fixedIDs{}).New("e1", assignments, shape, members)
	if err != nil {
		t.Fatal(err)
	}
	if err := runningOrder.Confirm(0); err != nil {
		t.Fatal(err)
	}
	return runningOrder
}

func TestNewValidatesLegs(t *testing.T) {
	factory := NewFactory(fixedIDs{})
	distance3 := Shape{LegCount: 3}
	time := Shape{TimeBased: true}

	tests := []struct {
		name        string
		shape       Shape
		assignments []Assignment
		wantErr     bool
	}{
		{name: "distance, all legs", shape: distance3, assignments: assign(runner("a"), nil, runner("a"))},
		{name: "distance, too few legs", shape: distance3, assignments: assign(runner("a")), wantErr: true},
		{name: "time, any number of legs", shape: time, assignments: assign(runner("a"), runner("b"), runner("a"), runner("c"))},
		{name: "time, no legs", shape: time, wantErr: true},
		{name: "gap in numbers", shape: time, assignments: []Assignment{{Number: 1}, {Number: 3}}, wantErr: true},
		{name: "runner not in the entry", shape: time, assignments: assign(runner("z")), wantErr: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := factory.New("e1", testCase.assignments, testCase.shape, members)
			if testCase.wantErr != isValidation(err) {
				t.Errorf("New() error = %v, want a ValidationError: %v", err, testCase.wantErr)
			}
		})
	}
}

func TestConfirm(t *testing.T) {
	runningOrder, err := NewFactory(fixedIDs{}).New("e1", assign(runner("a"), nil), Shape{LegCount: 2}, members)
	if err != nil {
		t.Fatal(err)
	}
	if err := runningOrder.Confirm(0); !isValidation(err) {
		t.Errorf("Confirm() with an empty leg = %v, want a ValidationError", err)
	}
	if _, err := runningOrder.Replace(assign(runner("a"), runner("b")), Shape{LegCount: 2}, members, 0); err != nil {
		t.Fatal(err)
	}
	if err := runningOrder.Confirm(0); !errors.Is(err, shared.ErrVersionConflict) {
		t.Errorf("Confirm(stale version) = %v, want ErrVersionConflict", err)
	}
	if err := runningOrder.Confirm(1); err != nil || runningOrder.Status() != Confirmed {
		t.Errorf("Confirm() = %v, status %s", err, runningOrder.Status())
	}
	var stateErr *shared.StateError
	if err := runningOrder.Confirm(2); !errors.As(err, &stateErr) {
		t.Errorf("Confirm() twice = %v, want a StateError", err)
	}
}

func TestReplaceLapLoss(t *testing.T) {
	shape := Shape{LegCount: 3}
	tests := []struct {
		name     string
		next     []Assignment
		wantLost []shared.MemberID
	}{
		{name: "moving a runner keeps laps", next: assign(runner("b"), runner("a"), runner("a")), wantLost: nil},
		{name: "fewer legs loses laps", next: assign(runner("a"), runner("b"), runner("b")), wantLost: []shared.MemberID{"a"}},
		{name: "removed runner loses laps", next: assign(runner("a"), runner("a"), nil), wantLost: []shared.MemberID{"b"}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			runningOrder := newConfirmed(t, shape, assign(runner("a"), runner("a"), runner("b")))
			lost, err := runningOrder.Replace(testCase.next, shape, members, runningOrder.Version())
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(lost, testCase.wantLost) {
				t.Errorf("lost = %v, want %v", lost, testCase.wantLost)
			}
			if runningOrder.Status() != Confirmed {
				t.Error("a confirmed running order must stay confirmed")
			}
		})
	}

	draft, _ := NewFactory(fixedIDs{}).New("e1", assign(runner("a"), runner("a"), runner("b")), shape, members)
	if lost, _ := draft.Replace(assign(nil, nil, nil), shape, members, 0); lost != nil {
		t.Errorf("a draft never loses laps, got %v", lost)
	}
}

func TestAdjustTo(t *testing.T) {
	runningOrder := newConfirmed(t, Shape{LegCount: 3}, assign(runner("a"), runner("b"), runner("c")))
	lost := runningOrder.AdjustTo(Shape{LegCount: 2})
	if got := runners(runningOrder); !slices.Equal(got, []string{"a", "b"}) || !slices.Equal(lost, []shared.MemberID{"c"}) {
		t.Errorf("fewer legs: legs %v, lost %v", got, lost)
	}
	lost = runningOrder.AdjustTo(Shape{LegCount: 4})
	if got := runners(runningOrder); !slices.Equal(got, []string{"a", "b", "-", "-"}) || lost != nil {
		t.Errorf("more legs: legs %v, lost %v", got, lost)
	}
	if lost := runningOrder.AdjustTo(Shape{TimeBased: true}); lost != nil || len(runningOrder.Legs()) != 4 {
		t.Error("distance → time keeps the legs")
	}
	if !runningOrder.HasEmptyLegs() || runningOrder.Status() != Confirmed {
		t.Error("the order stays confirmed with empty legs")
	}
}

func TestRemoveRunner(t *testing.T) {
	runningOrder := newConfirmed(t, Shape{TimeBased: true}, assign(runner("a"), runner("b"), runner("a")))
	lost := runningOrder.RemoveRunner("a")
	if got := runners(runningOrder); !slices.Equal(got, []string{"-", "b", "-"}) || !slices.Equal(lost, []shared.MemberID{"a"}) {
		t.Errorf("legs %v, lost %v", got, lost)
	}
	if lost := runningOrder.RemoveRunner("z"); lost != nil {
		t.Errorf("removing a non-runner: lost %v", lost)
	}
}

func TestLoadedVersion(t *testing.T) {
	runningOrder := NewFactory(fixedIDs{}).Reconstruct("o1", "e1", Confirmed, 4, assign(runner("a"), runner("b")))
	runningOrder.RemoveRunner("a")
	runningOrder.AdjustTo(Shape{LegCount: 3})
	if runningOrder.Version() != 6 || runningOrder.LoadedVersion() != 4 {
		t.Errorf("Version() = %d, LoadedVersion() = %d, want 6 and 4", runningOrder.Version(), runningOrder.LoadedVersion())
	}
}

func TestVisibilityPolicy(t *testing.T) {
	policy := VisibilityPolicy{}
	if !policy.CanSee(Viewer{Organizer: true}, Draft) {
		t.Error("organizers see drafts")
	}
	if policy.CanSee(Viewer{Attending: true}, Draft) {
		t.Error("members do not see drafts")
	}
	if !policy.CanSee(Viewer{Attending: true}, Confirmed) || policy.CanSee(Viewer{}, Confirmed) {
		t.Error("only attending members see confirmed orders")
	}
}
