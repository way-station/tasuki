package participations

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestCancelInvitationUseCase(t *testing.T) {
	cancel := func(t *testing.T, invitations *fakeInvitationRepository, entries *fakeEntryRepository, runners *fakeRunnerRemover) error {
		t.Helper()
		command := ParticipationCommandFactory{}.Invitation("r1", "m1")
		races := &fakeRaceRepository{stored: teamRace(t, teamID)}
		return NewCancelInvitationUseCase(&fakeUnitOfWork{}, races, invitations, entries, runners, teamID).Execute(context.Background(), command)
	}

	t.Run("takes an entry member out of the entry and the running orders", func(t *testing.T) {
		invitations, entries, runners := invitedTo("r1", "m1"), entriesOf(entry("r1", "m1", "m2")), &fakeRunnerRemover{}
		if err := cancel(t, invitations, entries, runners); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if invitations.deleted == nil || *invitations.deleted != (invitationKey{"r1", "m1"}) || !invitations.deleteInTransaction {
			t.Errorf("deleted %+v in transaction %v, want m1 to r1 in the transaction", invitations.deleted, invitations.deleteInTransaction)
		}
		if len(entries.saved) != 1 || entries.saved[0].Has("m1") || !entries.saved[0].Has("m2") || !entries.savedInTransaction {
			t.Errorf("saved entries %+v, want the entry without m1 in the transaction", entries.saved)
		}
		if len(runners.removed) != 1 || runners.removed[0] != (removedRunner{raceID: "r1", memberID: "m1", inTransaction: true}) {
			t.Errorf("removed runners %+v, want m1 of r1 in the transaction", runners.removed)
		}
	})

	t.Run("leaves the running orders of a member not in the entry", func(t *testing.T) {
		invitations, entries, runners := invitedTo("r1", "m1"), entriesOf(entry("r1", "m2")), &fakeRunnerRemover{}
		if err := cancel(t, invitations, entries, runners); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if invitations.deleted == nil {
			t.Error("the invitation was not deleted")
		}
		if len(entries.saved) != 0 || len(runners.removed) != 0 {
			t.Errorf("saved entries %+v, removed runners %+v, want neither", entries.saved, runners.removed)
		}
	})

	t.Run("does nothing for a member who is not invited", func(t *testing.T) {
		invitations, entries, runners := invitedTo("r1"), entriesOf(entry("r1", "m1")), &fakeRunnerRemover{}
		if err := cancel(t, invitations, entries, runners); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if invitations.deleted != nil || len(entries.saved) != 0 || len(runners.removed) != 0 {
			t.Error("something was changed for a member who is not invited")
		}
	})

	t.Run("rolls back when the running orders cannot be changed", func(t *testing.T) {
		unitOfWork := &fakeUnitOfWork{}
		races := &fakeRaceRepository{stored: teamRace(t, teamID)}
		invitations := invitedTo("r1", "m1")
		runners := &fakeRunnerRemover{err: errDatabase}
		command := ParticipationCommandFactory{}.Invitation("r1", "m1")
		err := NewCancelInvitationUseCase(unitOfWork, races, invitations, entriesOf(entry("r1", "m1")), runners, teamID).Execute(context.Background(), command)
		if !errors.Is(err, errDatabase) || !errors.Is(unitOfWork.err, errDatabase) {
			t.Errorf("Execute() error = %v, unit of work got %v, want %v", err, unitOfWork.err, errDatabase)
		}
		if invitations.deleted != nil {
			t.Error("the invitation was deleted")
		}
	})

	t.Run("returns not found for a race of another team", func(t *testing.T) {
		races := &fakeRaceRepository{stored: teamRace(t, "t2")}
		invitations := invitedTo("r1", "m1")
		command := ParticipationCommandFactory{}.Invitation("r1", "m1")
		err := NewCancelInvitationUseCase(&fakeUnitOfWork{}, races, invitations, entriesOf(), &fakeRunnerRemover{}, teamID).Execute(context.Background(), command)
		if !errors.Is(err, shared.ErrNotFound) || invitations.deleted != nil {
			t.Errorf("Execute() error = %v, deleted %+v, want not found and nothing deleted", err, invitations.deleted)
		}
	})
}
