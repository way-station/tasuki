package rosters

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestUnlinkUseCase(t *testing.T) {
	t.Run("unlinks the approved link", func(t *testing.T) {
		links := linksOf(link("k1", "u1", "m1", true, false))
		if err := NewUnlinkUseCase(fakeUnitOfWork{}, membersOf(teamMember("m1", teamID)), links, clock, teamID).Execute(context.Background(), "m1"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !links.stored["k1"].Unlinked() {
			t.Error("the link was not unlinked")
		}
	})

	t.Run("rejects a member without an approved link", func(t *testing.T) {
		links := linksOf(link("k1", "u1", "m1", false, false))
		err := NewUnlinkUseCase(fakeUnitOfWork{}, membersOf(teamMember("m1", teamID)), links, clock, teamID).Execute(context.Background(), "m1")
		var stateErr *shared.StateError
		if !errors.As(err, &stateErr) || len(links.saved) != 0 {
			t.Errorf("Execute() error = %v, want a state error and nothing saved", err)
		}
	})

	t.Run("rejects a member of another team", func(t *testing.T) {
		err := NewUnlinkUseCase(fakeUnitOfWork{}, membersOf(teamMember("m1", "t2")), linksOf(), clock, teamID).Execute(context.Background(), "m1")
		if !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute() error = %v, want shared.ErrNotFound", err)
		}
	})
}
