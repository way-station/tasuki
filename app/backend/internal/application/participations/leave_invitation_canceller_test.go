package participations

import (
	"context"
	"testing"
)

func TestLeaveInvitationCanceller(t *testing.T) {
	t.Run("cancels the invitation of the race with its side effects", func(t *testing.T) {
		races := &fakeRaceRepository{stored: teamRace(t, teamID)}
		invitations, entries, runners := invitedTo("r1", "m1"), entriesOf(entry("r1", "m1", "m2")), &fakeRunnerRemover{}
		cancel := NewCancelInvitationUseCase(&fakeUnitOfWork{}, races, invitations, entries, runners, teamID)
		if err := NewLeaveInvitationCanceller(cancel, ParticipationCommandFactory{}).CancelInvitation(context.Background(), "r1", "m1"); err != nil {
			t.Fatalf("CancelInvitation() error = %v", err)
		}
		if invitations.deleted == nil || *invitations.deleted != (invitationKey{"r1", "m1"}) {
			t.Errorf("deleted %+v, want m1 to r1", invitations.deleted)
		}
		if len(entries.saved) != 1 || entries.saved[0].Has("m1") {
			t.Errorf("saved entries %+v, want the entry without m1", entries.saved)
		}
		if len(runners.removed) != 1 || runners.removed[0].memberID != "m1" || runners.removed[0].raceID != "r1" {
			t.Errorf("removed runners %+v, want m1 of r1", runners.removed)
		}
	})
}
