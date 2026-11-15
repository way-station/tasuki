package rosters

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestDecideLinkUseCase(t *testing.T) {
	newUseCase := func(links *fakeLinkRepository, team shared.TeamID) *DecideLinkUseCase {
		return NewDecideLinkUseCase(fakeUnitOfWork{}, membersOf(teamMember("m1", team)), links, clock, teamID)
	}

	t.Run("approves a pending link", func(t *testing.T) {
		links := linksOf(link("k1", "u1", "m1", false, false))
		if err := newUseCase(links, teamID).Approve(context.Background(), "k1"); err != nil || !links.stored["k1"].Active() {
			t.Errorf("Approve() error = %v, active = %v, want an approved link", err, links.stored["k1"].Active())
		}
	})

	t.Run("rejects a pending link", func(t *testing.T) {
		links := linksOf(link("k1", "u1", "m1", false, false))
		if err := newUseCase(links, teamID).Reject(context.Background(), "k1"); err != nil || !links.stored["k1"].Rejected() {
			t.Errorf("Reject() error = %v, want a rejected link", err)
		}
	})

	t.Run("refuses a link that does not wait for approval", func(t *testing.T) {
		links := linksOf(link("k1", "u1", "m1", true, false))
		var stateErr *shared.StateError
		if err := newUseCase(links, teamID).Reject(context.Background(), "k1"); !errors.As(err, &stateErr) || len(links.saved) != 0 {
			t.Errorf("Reject() error = %v, want a state error and nothing saved", err)
		}
	})

	t.Run("refuses an unknown link and a link of another team", func(t *testing.T) {
		if err := newUseCase(linksOf(), teamID).Approve(context.Background(), "k9"); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Approve(k9) error = %v, want shared.ErrNotFound", err)
		}
		if err := newUseCase(linksOf(link("k1", "u1", "m1", false, false)), "t2").Approve(context.Background(), "k1"); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Approve() of another team error = %v, want shared.ErrNotFound", err)
		}
	})
}
