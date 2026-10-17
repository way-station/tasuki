package participation

import (
	"errors"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "e1" }

func isValidation(err error) bool { var ve *shared.ValidationError; return errors.As(err, &ve) }

func TestInvitationFactory(t *testing.T) {
	if _, err := (InvitationFactory{}).New("r1", "m1", false); err != nil {
		t.Errorf("New() = %v", err)
	}
	if _, err := (InvitationFactory{}).New("r1", "m1", true); !isValidation(err) {
		t.Errorf("New(left member) = %v, want a ValidationError", err)
	}
}

func TestAttendanceChange(t *testing.T) {
	inv := InvitationFactory{}.Reconstruct("r1", "m1")
	a, err := AttendanceFactory{}.New(inv, Attending)
	if err != nil {
		t.Fatal(err)
	}

	left, err := a.Change(Undecided)
	if err != nil || !left {
		t.Errorf("attending → undecided: left = %v, err = %v, want left", left, err)
	}
	left, err = a.Change(Attending)
	if err != nil || left {
		t.Errorf("undecided → attending: left = %v, err = %v", left, err)
	}
	if _, err := a.Change("maybe"); !isValidation(err) {
		t.Errorf("Change(unknown) = %v, want a ValidationError", err)
	}
	left, err = a.Change(NotAttending)
	if err != nil || !left {
		t.Errorf("attending → not attending: left = %v, err = %v, want left", left, err)
	}
	if _, err := a.Change(Attending); !errors.Is(err, ErrAnswerFinal) {
		t.Errorf("after not attending: err = %v, want ErrAnswerFinal", err)
	}
}

func TestEntry(t *testing.T) {
	f := NewEntryFactory(fixedIDs{})
	if _, err := f.New("r1", "Team A", 1); err == nil {
		t.Error("New() with an existing entry: want an error")
	}
	e, err := f.New("r1", " Team A ", 0)
	if err != nil {
		t.Fatal(err)
	}
	if e.Name() != "Team A" {
		t.Errorf("Name() = %q", e.Name())
	}

	if err := e.Add("m1", Undecided); !isValidation(err) {
		t.Errorf("Add(undecided) = %v, want a ValidationError", err)
	}
	if err := e.Add("m1", Attending); err != nil {
		t.Fatal(err)
	}
	_ = e.Add("m1", Attending) // does nothing
	_ = e.Add("m2", Attending)
	if got := e.Members(); !slices.Equal(got, []shared.MemberID{"m1", "m2"}) {
		t.Errorf("Members() = %v", got)
	}
	if !e.Remove("m1") || e.Remove("m1") || e.Has("m1") {
		t.Error("Remove() is wrong")
	}
}

func TestAttendancePolicy(t *testing.T) {
	p := AttendancePolicy{}
	if !p.CanSeeAllAnswers(true) || p.CanSeeAllAnswers(false) {
		t.Error("CanSeeAllAnswers() is wrong")
	}
}
