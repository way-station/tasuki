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
	as := make([]Assignment, len(runners))
	for i, r := range runners {
		as[i] = Assignment{Number: i + 1, Runner: r}
	}
	return as
}

func runners(o *RunningOrder) []string {
	var got []string
	for _, l := range o.Legs() {
		r, ok := l.Runner()
		if !ok {
			r = "-"
		}
		got = append(got, string(r))
	}
	return got
}

func isValidation(err error) bool { var ve *shared.ValidationError; return errors.As(err, &ve) }

func newConfirmed(t *testing.T, shape Shape, as []Assignment) *RunningOrder {
	t.Helper()
	o, err := NewFactory(fixedIDs{}).New("e1", as, shape, members)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Confirm(0); err != nil {
		t.Fatal(err)
	}
	return o
}

func TestNewValidatesLegs(t *testing.T) {
	f := NewFactory(fixedIDs{})
	distance3 := Shape{LegCount: 3}
	time := Shape{TimeBased: true}

	tests := []struct {
		name    string
		shape   Shape
		as      []Assignment
		wantErr bool
	}{
		{name: "distance, all legs", shape: distance3, as: assign(runner("a"), nil, runner("a"))},
		{name: "distance, too few legs", shape: distance3, as: assign(runner("a")), wantErr: true},
		{name: "time, any number of legs", shape: time, as: assign(runner("a"), runner("b"), runner("a"), runner("c"))},
		{name: "time, no legs", shape: time, wantErr: true},
		{name: "gap in numbers", shape: time, as: []Assignment{{Number: 1}, {Number: 3}}, wantErr: true},
		{name: "runner not in the entry", shape: time, as: assign(runner("z")), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.New("e1", tt.as, tt.shape, members)
			if tt.wantErr != isValidation(err) {
				t.Errorf("New() error = %v, want a ValidationError: %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfirm(t *testing.T) {
	o, err := NewFactory(fixedIDs{}).New("e1", assign(runner("a"), nil), Shape{LegCount: 2}, members)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Confirm(0); !isValidation(err) {
		t.Errorf("Confirm() with an empty leg = %v, want a ValidationError", err)
	}
	if _, err := o.Replace(assign(runner("a"), runner("b")), Shape{LegCount: 2}, members, 0); err != nil {
		t.Fatal(err)
	}
	if err := o.Confirm(0); !errors.Is(err, shared.ErrVersionConflict) {
		t.Errorf("Confirm(stale version) = %v, want ErrVersionConflict", err)
	}
	if err := o.Confirm(1); err != nil || o.Status() != Confirmed {
		t.Errorf("Confirm() = %v, status %s", err, o.Status())
	}
	var se *shared.StateError
	if err := o.Confirm(2); !errors.As(err, &se) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := newConfirmed(t, shape, assign(runner("a"), runner("a"), runner("b")))
			lost, err := o.Replace(tt.next, shape, members, o.Version())
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(lost, tt.wantLost) {
				t.Errorf("lost = %v, want %v", lost, tt.wantLost)
			}
			if o.Status() != Confirmed {
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
	o := newConfirmed(t, Shape{LegCount: 3}, assign(runner("a"), runner("b"), runner("c")))
	lost := o.AdjustTo(Shape{LegCount: 2})
	if got := runners(o); !slices.Equal(got, []string{"a", "b"}) || !slices.Equal(lost, []shared.MemberID{"c"}) {
		t.Errorf("fewer legs: legs %v, lost %v", got, lost)
	}
	lost = o.AdjustTo(Shape{LegCount: 4})
	if got := runners(o); !slices.Equal(got, []string{"a", "b", "-", "-"}) || lost != nil {
		t.Errorf("more legs: legs %v, lost %v", got, lost)
	}
	if lost := o.AdjustTo(Shape{TimeBased: true}); lost != nil || len(o.Legs()) != 4 {
		t.Error("distance → time keeps the legs")
	}
	if !o.HasEmptyLegs() || o.Status() != Confirmed {
		t.Error("the order stays confirmed with empty legs")
	}
}

func TestRemoveRunner(t *testing.T) {
	o := newConfirmed(t, Shape{TimeBased: true}, assign(runner("a"), runner("b"), runner("a")))
	lost := o.RemoveRunner("a")
	if got := runners(o); !slices.Equal(got, []string{"-", "b", "-"}) || !slices.Equal(lost, []shared.MemberID{"a"}) {
		t.Errorf("legs %v, lost %v", got, lost)
	}
	if lost := o.RemoveRunner("z"); lost != nil {
		t.Errorf("removing a non-runner: lost %v", lost)
	}
}

func TestLoadedVersion(t *testing.T) {
	o := NewFactory(fixedIDs{}).Reconstruct("o1", "e1", Confirmed, 4, assign(runner("a"), runner("b")))
	o.RemoveRunner("a")
	o.AdjustTo(Shape{LegCount: 3})
	if o.Version() != 6 || o.LoadedVersion() != 4 {
		t.Errorf("Version() = %d, LoadedVersion() = %d, want 6 and 4", o.Version(), o.LoadedVersion())
	}
}

func TestVisibilityPolicy(t *testing.T) {
	p := VisibilityPolicy{}
	if !p.CanSee(Viewer{Organizer: true}, Draft) {
		t.Error("organizers see drafts")
	}
	if p.CanSee(Viewer{Attending: true}, Draft) {
		t.Error("members do not see drafts")
	}
	if !p.CanSee(Viewer{Attending: true}, Confirmed) || p.CanSee(Viewer{}, Confirmed) {
		t.Error("only attending members see confirmed orders")
	}
}
