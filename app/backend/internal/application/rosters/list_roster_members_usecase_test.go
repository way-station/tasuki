package rosters

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const teamID shared.TeamID = "t1"

var errDatabase = errors.New("database is down")

var testNow = time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)

func clock() time.Time { return testNow }

func at(offset time.Duration) *time.Time {
	moment := testNow.Add(offset)
	return &moment
}

type fakeRosterQueryService struct {
	members []RosterMemberRecord
	detail  RosterMemberDetailRecord
	pending []PendingLinkRecord
	invited []roster.InvitedRace
	err     error
}

func (fake *fakeRosterQueryService) ListRosterMembers(_ context.Context, team shared.TeamID) ([]RosterMemberRecord, error) {
	if team != teamID {
		return nil, fmt.Errorf("team %s was not expected", team)
	}
	return fake.members, fake.err
}

func (fake *fakeRosterQueryService) GetRosterMember(_ context.Context, team shared.TeamID, memberID shared.MemberID) (RosterMemberDetailRecord, error) {
	if fake.err != nil {
		return RosterMemberDetailRecord{}, fake.err
	}
	if team != teamID || memberID != fake.detail.MemberID {
		return RosterMemberDetailRecord{}, fmt.Errorf("roster member %s: %w", memberID, shared.ErrNotFound)
	}
	return fake.detail, nil
}

func (fake *fakeRosterQueryService) ListPendingLinks(_ context.Context, team shared.TeamID) ([]PendingLinkRecord, error) {
	if team != teamID {
		return nil, fmt.Errorf("team %s was not expected", team)
	}
	return fake.pending, fake.err
}

func (fake *fakeRosterQueryService) ListInvitedRaces(context.Context, shared.MemberID) ([]roster.InvitedRace, error) {
	return fake.invited, fake.err
}

func newListUseCase(query *fakeRosterQueryService) *ListRosterMembersUseCase {
	return NewListRosterMembersUseCase(query, identity.NewMemberLinkFactory(nil), identity.NewInviteLinkFactory(nil, nil), identity.LinkStatusService{}, RosterDTOFactory{}, clock, teamID)
}

func TestListRosterMembersUseCase(t *testing.T) {
	query := &fakeRosterQueryService{members: []RosterMemberRecord{
		{MemberID: "m1", Name: "Hanako", Links: []LinkRecord{{ID: "k1", UserID: "u1", CreatedAt: testNow, ApprovedAt: at(time.Hour)}}},
		{MemberID: "m2", Name: "Jiro", Links: []LinkRecord{{ID: "k2", UserID: "u2", CreatedAt: testNow}}},
		{MemberID: "m3", Name: "Saburo", Left: true, Links: []LinkRecord{{ID: "k3", UserID: "u3", CreatedAt: testNow, RejectedAt: at(time.Hour)}}},
		{MemberID: "m4", Name: "Taro", InviteLink: &InviteLinkRecord{ID: "i4", ExpiresAt: testNow.Add(24 * time.Hour)}},
		{MemberID: "m5", Name: "Yuki", InviteLink: &InviteLinkRecord{ID: "i5", ExpiresAt: testNow.Add(-time.Hour)}},
	}}

	members, err := newListUseCase(query).Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := []struct {
		status    roster.LinkStatus
		left      bool
		expiresAt bool
	}{{roster.LinkApproved, false, false}, {roster.LinkPending, false, false}, {roster.LinkNone, true, false}, {roster.LinkNone, false, true}, {roster.LinkNone, false, false}}
	if len(members) != len(want) {
		t.Fatalf("Execute() = %d members, want %d", len(members), len(want))
	}
	for i, member := range members {
		_, hasExpiry := member.InviteLinkExpiresAt()
		if member.LinkStatus() != want[i].status || member.Left() != want[i].left || hasExpiry != want[i].expiresAt {
			t.Errorf("members[%d] = %+v, want %+v", i, member, want[i])
		}
	}

	t.Run("finds one member", func(t *testing.T) {
		member, err := newListUseCase(query).Find(context.Background(), "m4")
		if err != nil || member.Name() != "Taro" {
			t.Errorf("Find() = %+v, %v, want Taro", member, err)
		}
		if _, err := newListUseCase(query).Find(context.Background(), "m9"); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Find(m9) error = %v, want shared.ErrNotFound", err)
		}
	})

	t.Run("returns the error of the query service", func(t *testing.T) {
		if _, err := newListUseCase(&fakeRosterQueryService{err: errDatabase}).Execute(context.Background()); !errors.Is(err, errDatabase) {
			t.Errorf("Execute() error = %v, want %v", err, errDatabase)
		}
	})
}
