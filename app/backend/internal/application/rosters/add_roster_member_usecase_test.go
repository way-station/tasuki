package rosters

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type sequentialIDs struct{ next int }

func (ids *sequentialIDs) NewID() string {
	ids.next++
	return fmt.Sprintf("id%d", ids.next)
}

// fakeMemberRepository keeps roster members by ID.
type fakeMemberRepository struct {
	stored map[shared.MemberID]*roster.Member
	saved  []*roster.Member
}

func membersOf(members ...*roster.Member) *fakeMemberRepository {
	fake := &fakeMemberRepository{stored: map[shared.MemberID]*roster.Member{}}
	for _, member := range members {
		fake.stored[member.ID()] = member
	}
	return fake
}

func (fake *fakeMemberRepository) FindByID(_ context.Context, id shared.MemberID) (*roster.Member, error) {
	member, ok := fake.stored[id]
	if !ok {
		return nil, fmt.Errorf("roster member %s: %w", id, shared.ErrNotFound)
	}
	return member, nil
}

func (fake *fakeMemberRepository) Save(_ context.Context, member *roster.Member) error {
	fake.stored[member.ID()] = member
	fake.saved = append(fake.saved, member)
	return nil
}

func TestAddRosterMemberUseCase(t *testing.T) {
	t.Run("adds the member without a link", func(t *testing.T) {
		members := membersOf()
		added, err := NewAddRosterMemberUseCase(members, roster.NewMemberFactory(&sequentialIDs{}), RosterDTOFactory{}, teamID).Execute(context.Background(), RosterCommandFactory{}.AddRosterMember("Taro"))
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if added.MemberID() != "id1" || added.Name() != "Taro" || added.LinkStatus() != roster.LinkNone || len(members.saved) != 1 || members.saved[0].TeamID() != teamID {
			t.Errorf("Execute() = %+v, saved %v, want Taro of the team", added, members.saved)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		members := membersOf()
		_, err := NewAddRosterMemberUseCase(members, roster.NewMemberFactory(&sequentialIDs{}), RosterDTOFactory{}, teamID).Execute(context.Background(), RosterCommandFactory{}.AddRosterMember(" "))
		var validationErr *shared.ValidationError
		if !errors.As(err, &validationErr) || len(members.saved) != 0 {
			t.Errorf("Execute() error = %v, want a validation error and nothing saved", err)
		}
	})
}
