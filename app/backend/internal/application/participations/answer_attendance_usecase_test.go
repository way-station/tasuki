package participations

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestAnswerAttendanceUseCase(t *testing.T) {
	type answers = map[shared.MemberID]participation.AttendanceStatus
	answer := func(caller auth.MeDTO, invitations *fakeInvitationRepository, attendances *fakeAttendanceRepository, entries *fakeEntryRepository, runners *fakeRunnerRemover, status string) (AttendanceDTO, error) {
		command := ParticipationCommandFactory{}.Attendance("r1", status)
		useCase := NewAnswerAttendanceUseCase(&fakeUnitOfWork{}, invitations, attendances, entries, runners, participation.AttendanceFactory{}, ParticipationDTOFactory{})
		return useCase.Execute(context.Background(), caller, command)
	}

	tests := []struct {
		name       string
		before     answers
		status     string
		leaveEntry bool
	}{
		{name: "the first answer", before: answers{}, status: "attending"},
		{name: "attending to undecided", before: answers{"m1": participation.Attending}, status: "undecided", leaveEntry: true},
		{name: "attending to not attending", before: answers{"m1": participation.Attending}, status: "not_attending", leaveEntry: true},
		{name: "undecided to attending", before: answers{"m1": participation.Undecided}, status: "attending"},
		{name: "attending again", before: answers{"m1": participation.Attending}, status: "attending"},
	}
	for _, testCase := range tests {
		t.Run("saves the answer: "+testCase.name, func(t *testing.T) {
			attendances, entries, runners := answered("r1", testCase.before), entriesOf(entry("r1", "m1")), &fakeRunnerRemover{}
			got, err := answer(member(), invitedTo("r1", "m1"), attendances, entries, runners, testCase.status)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			want := participation.AttendanceStatus(testCase.status)
			if got.Status() != want || attendances.saved == nil || attendances.saved.MemberID() != "m1" || attendances.saved.Status() != want || !attendances.savedInTransaction {
				t.Errorf("returned %q, saved %+v, want %q of m1 in the transaction", got.Status(), attendances.saved, want)
			}
			if left := len(entries.saved) == 1 && !entries.saved[0].Has("m1"); left != testCase.leaveEntry {
				t.Errorf("left the entry = %v, want %v", left, testCase.leaveEntry)
			}
			if removed := len(runners.removed) == 1 && runners.removed[0] == (removedRunner{raceID: "r1", memberID: "m1", inTransaction: true}); removed != testCase.leaveEntry {
				t.Errorf("removed runners %+v, want removed = %v", runners.removed, testCase.leaveEntry)
			}
		})
	}

	t.Run("rejects an unknown answer", func(t *testing.T) {
		attendances := answered("r1", answers{})
		if _, err := answer(member(), invitedTo("r1", "m1"), attendances, entriesOf(), &fakeRunnerRemover{}, "maybe"); !isValidation(err) {
			t.Errorf("Execute() error = %v, want a validation error", err)
		}
		if attendances.saved != nil {
			t.Error("an unknown answer was saved")
		}
	})

	notFound := []struct {
		name        string
		caller      auth.MeDTO
		invitations *fakeInvitationRepository
		before      answers
	}{
		{name: "an organizer without a link", caller: organizer(), invitations: invitedTo("r1", "m1"), before: answers{}},
		{name: "a member who is not invited", caller: member(), invitations: invitedTo("r1"), before: answers{}},
		{name: "a member who answered not attending", caller: member(), invitations: invitedTo("r1", "m1"), before: answers{"m1": participation.NotAttending}},
	}
	for _, testCase := range notFound {
		t.Run("returns not found for "+testCase.name, func(t *testing.T) {
			attendances := answered("r1", testCase.before)
			if _, err := answer(testCase.caller, testCase.invitations, attendances, entriesOf(), &fakeRunnerRemover{}, "attending"); !errors.Is(err, shared.ErrNotFound) {
				t.Errorf("Execute() error = %v, want %v", err, shared.ErrNotFound)
			}
			if attendances.saved != nil {
				t.Error("the answer was saved")
			}
		})
	}

	t.Run("rolls back when the running orders cannot be changed", func(t *testing.T) {
		unitOfWork := &fakeUnitOfWork{}
		command := ParticipationCommandFactory{}.Attendance("r1", "undecided")
		useCase := NewAnswerAttendanceUseCase(unitOfWork, invitedTo("r1", "m1"), answered("r1", answers{"m1": participation.Attending}), entriesOf(entry("r1", "m1")), &fakeRunnerRemover{err: errDatabase}, participation.AttendanceFactory{}, ParticipationDTOFactory{})
		if _, err := useCase.Execute(context.Background(), member(), command); !errors.Is(err, errDatabase) || !errors.Is(unitOfWork.err, errDatabase) {
			t.Errorf("Execute() error = %v, unit of work got %v, want %v", err, unitOfWork.err, errDatabase)
		}
	})
}
