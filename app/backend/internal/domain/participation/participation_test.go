package participation

import (
	"errors"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "e1" }

func isValidation(err error) bool {
	var validationErr *shared.ValidationError
	return errors.As(err, &validationErr)
}

func TestInvitationFactory(t *testing.T) {
	if _, err := (InvitationFactory{}).New("r1", "m1", false); err != nil {
		t.Errorf("New() = %v", err)
	}
	if _, err := (InvitationFactory{}).New("r1", "m1", true); !isValidation(err) {
		t.Errorf("New(left member) = %v, want attendance ValidationError", err)
	}
}

func TestAttendanceChange(t *testing.T) {
	inv := InvitationFactory{}.Reconstruct("r1", "m1")
	attendance, err := AttendanceFactory{}.New(inv, Attending)
	if err != nil {
		t.Fatal(err)
	}

	left, err := attendance.Change(Undecided)
	if err != nil || !left {
		t.Errorf("attending → undecided: left = %v, err = %v, want left", left, err)
	}
	left, err = attendance.Change(Attending)
	if err != nil || left {
		t.Errorf("undecided → attending: left = %v, err = %v", left, err)
	}
	if _, err := attendance.Change("maybe"); !isValidation(err) {
		t.Errorf("Change(unknown) = %v, want attendance ValidationError", err)
	}
	left, err = attendance.Change(NotAttending)
	if err != nil || !left {
		t.Errorf("attending → not attending: left = %v, err = %v, want left", left, err)
	}
	if _, err := attendance.Change(Attending); !errors.Is(err, ErrAnswerFinal) {
		t.Errorf("after not attending: err = %v, want ErrAnswerFinal", err)
	}
}

func TestEntry(t *testing.T) {
	factory := NewEntryFactory(fixedIDs{})
	if _, err := factory.New("r1", "Team A", 1); err == nil {
		t.Error("New() with an existing entry: want an error")
	}
	entry, err := factory.New("r1", " Team A ", 0)
	if err != nil {
		t.Fatal(err)
	}
	if entry.Name() != "Team A" {
		t.Errorf("Name() = %q", entry.Name())
	}

	if err := entry.Add("m1", Undecided); !isValidation(err) {
		t.Errorf("Add(undecided) = %v, want attendance ValidationError", err)
	}
	if err := entry.Add("m1", Attending); err != nil {
		t.Fatal(err)
	}
	_ = entry.Add("m1", Attending) // does nothing
	_ = entry.Add("m2", Attending)
	if got := entry.Members(); !slices.Equal(got, []shared.MemberID{"m1", "m2"}) {
		t.Errorf("Members() = %v", got)
	}
	if !entry.Remove("m1") || entry.Remove("m1") || entry.Has("m1") {
		t.Error("Remove() is wrong")
	}
}

func TestAttendancePolicy(t *testing.T) {
	policy := AttendancePolicy{}
	if !policy.CanSeeAllAnswers(true) || policy.CanSeeAllAnswers(false) {
		t.Error("CanSeeAllAnswers() is wrong")
	}
}
