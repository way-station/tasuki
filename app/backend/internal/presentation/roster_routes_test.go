package presentation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/application/rosters"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

const rosterTeamID shared.TeamID = "team-1"

var rosterNow = time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)

// rosterStore holds Taro (m1), Jiro (m2), and Saburo (m3) of the team, Goro (m9) of another
// team, their member links, and their invite links. Taro has a lap in race r1.
type rosterStore struct {
	members   map[shared.MemberID]*roster.Member
	links     map[shared.MemberLinkID]*identity.MemberLink
	invites   map[shared.InviteLinkID]*identity.InviteLink
	lineNames map[shared.UserID]string
}

func newRosterStore() *rosterStore {
	factory := roster.NewMemberFactory(nil)
	return &rosterStore{
		members: map[shared.MemberID]*roster.Member{
			"m1": factory.Reconstruct("m1", rosterTeamID, "Taro", nil, roster.DefaultAverage(), 0),
			"m2": factory.Reconstruct("m2", rosterTeamID, "Jiro", nil, roster.DefaultAverage(), 0),
			"m3": factory.Reconstruct("m3", rosterTeamID, "Saburo", nil, roster.DefaultAverage(), 0),
			"m9": factory.Reconstruct("m9", "team-2", "Goro", nil, roster.DefaultAverage(), 0),
		},
		links:     map[shared.MemberLinkID]*identity.MemberLink{},
		invites:   map[shared.InviteLinkID]*identity.InviteLink{},
		lineNames: map[shared.UserID]string{},
	}
}

// withLink links the member to a user named "LINE <name>", created an hour before
// rosterNow. An approved link was approved half an hour before rosterNow.
func (store *rosterStore) withLink(id shared.MemberLinkID, memberID shared.MemberID, userID shared.UserID, approved bool) *rosterStore {
	var approvedAt *time.Time
	if approved {
		approvedTime := rosterNow.Add(-30 * time.Minute)
		approvedAt = &approvedTime
	}
	store.links[id] = identity.NewMemberLinkFactory(nil).Reconstruct(id, userID, memberID, "i0", rosterNow.Add(-time.Hour), approvedAt, nil, nil)
	store.lineNames[userID] = "LINE " + store.members[memberID].Name()
	return store
}

// withInvite gives the member an open invite link that expires two days after rosterNow.
func (store *rosterStore) withInvite(id shared.InviteLinkID, memberID shared.MemberID) *rosterStore {
	store.invites[id] = identity.NewInviteLinkFactory(nil, nil).Reconstruct(id, memberID, nil, rosterNow.Add(48*time.Hour), nil, nil)
	return store
}

// withLeft marks the member as left a day before rosterNow.
func (store *rosterStore) withLeft(memberID shared.MemberID) *rosterStore {
	store.members[memberID].Leave(rosterNow.Add(-24 * time.Hour))
	return store
}

type storeRosterMembers struct{ store *rosterStore }

func (members storeRosterMembers) FindByID(_ context.Context, id shared.MemberID) (*roster.Member, error) {
	member, ok := members.store.members[id]
	if !ok {
		return nil, fmt.Errorf("member %s: %w", id, shared.ErrNotFound)
	}
	return member, nil
}

func (members storeRosterMembers) Save(_ context.Context, member *roster.Member) error {
	members.store.members[member.ID()] = member
	return nil
}

type storeMemberLinks struct{ store *rosterStore }

func (links storeMemberLinks) FindByID(_ context.Context, id shared.MemberLinkID) (*identity.MemberLink, error) {
	link, ok := links.store.links[id]
	if !ok {
		return nil, fmt.Errorf("member link %s: %w", id, shared.ErrNotFound)
	}
	return link, nil
}

func (links storeMemberLinks) ListByUser(_ context.Context, userID shared.UserID) ([]*identity.MemberLink, error) {
	var found []*identity.MemberLink
	for _, link := range links.store.links {
		if link.UserID() == userID {
			found = append(found, link)
		}
	}
	return found, nil
}

func (links storeMemberLinks) ListByMember(_ context.Context, memberID shared.MemberID) ([]*identity.MemberLink, error) {
	var found []*identity.MemberLink
	for _, link := range links.store.links {
		if link.MemberID() == memberID {
			found = append(found, link)
		}
	}
	return found, nil
}

func (links storeMemberLinks) Save(_ context.Context, link *identity.MemberLink) error {
	links.store.links[link.ID()] = link
	return nil
}

type storeInviteLinks struct{ store *rosterStore }

func (storeInviteLinks) FindByTokenHash(context.Context, []byte) (*identity.InviteLink, error) {
	return nil, errors.New("invite links are not found by token in the roster routes")
}

func (invites storeInviteLinks) FindOpenByMember(_ context.Context, memberID shared.MemberID) (*identity.InviteLink, error) {
	if invite := invites.store.openInvite(memberID); invite != nil {
		return invite, nil
	}
	return nil, fmt.Errorf("open invite link of %s: %w", memberID, shared.ErrNotFound)
}

func (invites storeInviteLinks) Save(_ context.Context, invite *identity.InviteLink) error {
	invites.store.invites[invite.ID()] = invite
	return nil
}

func (store *rosterStore) openInvite(memberID shared.MemberID) *identity.InviteLink {
	for _, invite := range store.invites {
		_, used := invite.UsedAt()
		_, revoked := invite.RevokedAt()
		if invite.MemberID() == memberID && !used && !revoked {
			return invite
		}
	}
	return nil
}

// noSessions is a user without active sessions.
type noSessions struct{}

func (noSessions) ListActiveByUser(context.Context, shared.UserID, time.Time) ([]*identity.Session, error) {
	return nil, nil
}

func (noSessions) FindByTokenHash(context.Context, []byte) (*identity.Session, error) {
	return nil, errors.New("sessions are not found by token in the roster routes")
}

func (noSessions) Save(context.Context, *identity.Session) error { return nil }

// noInvitations is a member who is not invited to any race, so nothing is cancelled.
type noInvitations struct{}

func (noInvitations) CancelInvitation(context.Context, shared.RaceID, shared.MemberID) error {
	return errors.New("no invitation should be cancelled")
}

type storeRosterQueries struct{ store *rosterStore }

func (queries storeRosterQueries) ListRosterMembers(_ context.Context, teamID shared.TeamID) ([]rosters.RosterMemberRecord, error) {
	var records []rosters.RosterMemberRecord
	for _, member := range queries.store.members {
		if member.TeamID() != teamID {
			continue
		}
		record := rosters.RosterMemberRecord{MemberID: member.ID(), Name: member.Name(), Left: member.Left()}
		for _, link := range queries.store.links {
			if link.MemberID() == member.ID() {
				record.Links = append(record.Links, linkRecordOf(link))
			}
		}
		if invite := queries.store.openInvite(member.ID()); invite != nil {
			record.InviteLink = &rosters.InviteLinkRecord{ID: invite.ID(), ExpiresAt: invite.ExpiresAt()}
		}
		records = append(records, record)
	}
	slices.SortFunc(records, func(left, right rosters.RosterMemberRecord) int { return strings.Compare(left.Name, right.Name) })
	return records, nil
}

func linkRecordOf(link *identity.MemberLink) rosters.LinkRecord {
	optional := func(moment time.Time, ok bool) *time.Time {
		if !ok {
			return nil
		}
		return &moment
	}
	return rosters.LinkRecord{
		ID: link.ID(), UserID: link.UserID(), InviteLinkID: link.InviteLinkID(), CreatedAt: link.CreatedAt(),
		ApprovedAt: optional(link.ApprovedAt()), RejectedAt: optional(link.RejectedAt()), UnlinkedAt: optional(link.UnlinkedAt()),
	}
}

func (queries storeRosterQueries) GetRosterMember(_ context.Context, teamID shared.TeamID, memberID shared.MemberID) (rosters.RosterMemberDetailRecord, error) {
	member, ok := queries.store.members[memberID]
	if !ok || member.TeamID() != teamID {
		return rosters.RosterMemberDetailRecord{}, fmt.Errorf("member %s: %w", memberID, shared.ErrNotFound)
	}
	average, _ := shared.ParsePace("3:40")
	record := rosters.RosterMemberDetailRecord{MemberID: member.ID(), Name: member.Name(), Left: member.Left(), Average: average, LapsByRace: []rosters.RaceLapsRecord{}}
	if memberID == "m1" {
		lapTime, _ := shared.ParsePace("3:30")
		record.Average, record.LapCount = lapTime, 1
		record.LapsByRace = []rosters.RaceLapsRecord{{RaceID: "r1", RaceName: "Ekiden", Laps: []rosters.LapRecord{{ID: "l1", Time: lapTime}}}}
	}
	return record, nil
}

func (queries storeRosterQueries) ListPendingLinks(_ context.Context, teamID shared.TeamID) ([]rosters.PendingLinkRecord, error) {
	var records []rosters.PendingLinkRecord
	for _, link := range queries.store.links {
		member := queries.store.members[link.MemberID()]
		if link.Pending() && member.TeamID() == teamID {
			records = append(records, rosters.PendingLinkRecord{LinkID: link.ID(), LineDisplayName: queries.store.lineNames[link.UserID()], MemberID: member.ID(), MemberName: member.Name(), CreatedAt: link.CreatedAt()})
		}
	}
	return records, nil
}

func (storeRosterQueries) ListInvitedRaces(context.Context, shared.MemberID) ([]roster.InvitedRace, error) {
	return nil, nil
}

func newRosterRouter(store *rosterStore) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	now := func() time.Time { return rosterNow }
	members, links, queries := storeRosterMembers{store}, storeMemberLinks{store}, storeRosterQueries{store}
	inviteLinkFactory := identity.NewInviteLinkFactory(&sequentialIDs{}, fixedToken("token-1"))
	rosterController := controller.NewRosterController(
		rosters.NewListRosterMembersUseCase(queries, identity.NewMemberLinkFactory(nil), inviteLinkFactory, identity.LinkStatusService{}, rosters.RosterDTOFactory{}, now, rosterTeamID),
		rosters.NewAddRosterMemberUseCase(members, roster.NewMemberFactory(&sequentialIDs{}), rosters.RosterDTOFactory{}, rosterTeamID),
		rosters.NewGetRosterMemberUseCase(queries, rosters.RosterDTOFactory{}, rosterTeamID),
		rosters.NewSetLeftUseCase(passUnitOfWork{}, members, queries, noInvitations{}, links, noSessions{}, noSessions{}, roster.LeaveService{}, now, time.UTC, rosterTeamID),
		rosters.NewUnlinkUseCase(passUnitOfWork{}, members, links, now, rosterTeamID),
		rosters.NewCreateInviteLinkUseCase(passUnitOfWork{}, members, links, storeInviteLinks{store}, inviteLinkFactory, identity.LinkStatusService{}, rosters.RosterDTOFactory{}, now, rosterTeamID),
		rosters.NewListPendingLinksUseCase(queries, rosters.RosterDTOFactory{}, rosterTeamID),
		rosters.NewDecideLinkUseCase(passUnitOfWork{}, members, links, now, rosterTeamID),
		rosters.RosterCommandFactory{},
		testOrigin,
	)
	check := health.NewCheckUseCase(fakeHealth{}, fakeHealth{}, health.StatusDTOFactory{}, logger)
	return NewRouter(Controllers{Health: controller.NewHealthController(check), Roster: rosterController}, Options{
		Authenticate:   middleware.Authenticate(tokenAuthenticator{}, logger),
		ResolveCaller:  middleware.ResolveCaller(tokenCallers{}, logger),
		FrontendOrigin: testOrigin,
		Logger:         logger,
	})
}

func TestRosterRoutesCallers(t *testing.T) {
	linked := func() *rosterStore {
		return newRosterStore().withLink("k1", "m1", "u1", true).withLink("k2", "m2", "u2", false)
	}
	tests := []struct {
		method, path, body string
		organizerWants     int
	}{
		{method: http.MethodGet, path: "/api/v1/roster-members", organizerWants: 200},
		{method: http.MethodPost, path: "/api/v1/roster-members", body: `{"name":"Shiro"}`, organizerWants: 201},
		{method: http.MethodGet, path: "/api/v1/roster-members/m1", organizerWants: 200},
		{method: http.MethodPatch, path: "/api/v1/roster-members/m3", body: `{"left":true}`, organizerWants: 200},
		{method: http.MethodDelete, path: "/api/v1/roster-members/m1/link", organizerWants: 204},
		{method: http.MethodPost, path: "/api/v1/roster-members/m3/invite-links", organizerWants: 201},
		{method: http.MethodGet, path: "/api/v1/member-links", organizerWants: 200},
		{method: http.MethodPost, path: "/api/v1/member-links/k2/approve", organizerWants: 204},
		{method: http.MethodPost, path: "/api/v1/member-links/k2/reject", organizerWants: 204},
	}
	for _, testCase := range tests {
		// Every roster endpoint is for organizers only (docs/design/api.md).
		want := map[string]int{"": 401, waitingToken: 403, memberToken: 403, organizerToken: testCase.organizerWants}
		for _, token := range []string{"", waitingToken, memberToken, organizerToken} {
			t.Run(fmt.Sprintf("%s %s as %q", testCase.method, testCase.path, token), func(t *testing.T) {
				rec := serve(newRosterRouter(linked()), testCase.method, testCase.path, testCase.body, sessionOf(token))
				if rec.Code != want[token] {
					t.Errorf("status = %d, want %d (body %s)", rec.Code, want[token], rec.Body)
				}
			})
		}
	}
}

func TestRosterRoutesBodies(t *testing.T) {
	inviteExpiresAt := rosterNow.Add(identity.InviteLinkLifetime).Format(time.RFC3339)
	tests := []struct {
		name, method, path, body string
		store                    *rosterStore
		wantStatus               int
		want                     string
	}{
		{name: "lists the roster by name with the link states and the open invite link", method: http.MethodGet, path: "/api/v1/roster-members",
			store: newRosterStore().withLink("k1", "m1", "u1", true).withLink("k2", "m2", "u2", false).withInvite("i3", "m3"), wantStatus: 200,
			want: `{"roster_members":[` +
				`{"id":"m2","invite_link_expires_at":null,"left":false,"link_status":"pending","name":"Jiro"},` +
				`{"id":"m3","invite_link_expires_at":"2026-11-15T09:00:00Z","left":false,"link_status":"none","name":"Saburo"},` +
				`{"id":"m1","invite_link_expires_at":null,"left":false,"link_status":"approved","name":"Taro"}]}`},
		{name: "returns the added member", method: http.MethodPost, path: "/api/v1/roster-members", body: `{"name":"Shiro"}`,
			store: newRosterStore(), wantStatus: 201,
			want: `{"id":"id1","invite_link_expires_at":null,"left":false,"link_status":"none","name":"Shiro"}`},
		{name: "returns the profile with the laps by race", method: http.MethodGet, path: "/api/v1/roster-members/m1",
			store: newRosterStore(), wantStatus: 200,
			want: `{"average_time":"3:30","id":"m1","lap_count":1,"laps_by_race":[{"laps":[{"id":"l1","time":"3:30"}],"race_id":"r1","race_name":"Ekiden"}],"left":false,"name":"Taro"}`},
		{name: "returns an empty laps_by_race for a member without laps", method: http.MethodGet, path: "/api/v1/roster-members/m2",
			store: newRosterStore(), wantStatus: 200,
			want: `{"average_time":"3:40","id":"m2","lap_count":0,"laps_by_race":[],"left":false,"name":"Jiro"}`},
		{name: "returns the member marked as left", method: http.MethodPatch, path: "/api/v1/roster-members/m3", body: `{"left":true}`,
			store: newRosterStore(), wantStatus: 200,
			want: `{"id":"m3","invite_link_expires_at":null,"left":true,"link_status":"none","name":"Saburo"}`},
		{name: "returns the member back from leaving", method: http.MethodPatch, path: "/api/v1/roster-members/m3", body: `{"left":false}`,
			store: newRosterStore().withLeft("m3"), wantStatus: 200,
			want: `{"id":"m3","invite_link_expires_at":null,"left":false,"link_status":"none","name":"Saburo"}`},
		{name: "returns the invite link URL on the frontend's origin", method: http.MethodPost, path: "/api/v1/roster-members/m3/invite-links",
			store: newRosterStore().withInvite("i3", "m3"), wantStatus: 201,
			want: `{"expires_at":"` + inviteExpiresAt + `","url":"https://tasuki.example/invite/token-1"}`},
		{name: "lists the pending links of the team", method: http.MethodGet, path: "/api/v1/member-links",
			store: newRosterStore().withLink("k1", "m1", "u1", true).withLink("k2", "m2", "u2", false).withLink("k9", "m9", "u9", false), wantStatus: 200,
			want: `{"links":[{"created_at":"2026-11-13T08:00:00Z","id":"k2","line_display_name":"LINE Jiro","roster_member":{"id":"m2","name":"Jiro"}}]}`},
		{name: "lists no pending links as an empty list", method: http.MethodGet, path: "/api/v1/member-links",
			store: newRosterStore(), wantStatus: 200,
			want: `{"links":[]}`},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := serve(newRosterRouter(testCase.store), testCase.method, testCase.path, testCase.body, sessionOf(organizerToken))
			if rec.Code != testCase.wantStatus || rec.Body.String() != testCase.want+"\n" {
				t.Errorf("status %d, body %s\nwant %d %s", rec.Code, rec.Body, testCase.wantStatus, testCase.want)
			}
		})
	}
}

func TestRosterRoutesChanges(t *testing.T) {
	tests := []struct {
		name, method, path string
		check              func(store *rosterStore) bool
	}{
		{name: "unlinking removes the approved link", method: http.MethodDelete, path: "/api/v1/roster-members/m1/link",
			check: func(store *rosterStore) bool { return store.links["k1"].Unlinked() }},
		{name: "approving approves the link", method: http.MethodPost, path: "/api/v1/member-links/k2/approve",
			check: func(store *rosterStore) bool { return store.links["k2"].Active() }},
		{name: "rejecting rejects the link", method: http.MethodPost, path: "/api/v1/member-links/k2/reject",
			check: func(store *rosterStore) bool { return store.links["k2"].Rejected() }},
		{name: "a new invite link replaces the previous one", method: http.MethodPost, path: "/api/v1/roster-members/m3/invite-links",
			check: func(store *rosterStore) bool {
				_, revoked := store.invites["i3"].RevokedAt()
				return revoked && store.openInvite("m3") != nil
			}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			store := newRosterStore().withLink("k1", "m1", "u1", true).withLink("k2", "m2", "u2", false).withInvite("i3", "m3")
			rec := serve(newRosterRouter(store), testCase.method, testCase.path, "", sessionOf(organizerToken))
			if rec.Code >= 300 || !testCase.check(store) {
				t.Errorf("status %d, body %s; the store did not change as expected", rec.Code, rec.Body)
			}
		})
	}
}

func TestRosterRoutesErrors(t *testing.T) {
	tests := []struct {
		name, method, path, body string
		wantStatus               int
		wantCode                 string
	}{
		{name: "an empty name", method: http.MethodPost, path: "/api/v1/roster-members", body: `{"name":""}`, wantStatus: 400, wantCode: "invalid_request"},
		{name: "a member without a body", method: http.MethodPost, path: "/api/v1/roster-members", wantStatus: 400, wantCode: "invalid_request"},
		{name: "a malformed body", method: http.MethodPost, path: "/api/v1/roster-members", body: `{"name":`, wantStatus: 400, wantCode: "invalid_request"},
		{name: "an update without a body", method: http.MethodPatch, path: "/api/v1/roster-members/m3", wantStatus: 400, wantCode: "invalid_request"},
		{name: "the profile of a member of another team", method: http.MethodGet, path: "/api/v1/roster-members/m9", wantStatus: 404, wantCode: "not_found"},
		{name: "the profile of an unknown member", method: http.MethodGet, path: "/api/v1/roster-members/m0", wantStatus: 404, wantCode: "not_found"},
		{name: "leaving for a member of another team", method: http.MethodPatch, path: "/api/v1/roster-members/m9", body: `{"left":true}`, wantStatus: 404, wantCode: "not_found"},
		{name: "unlinking a member of another team", method: http.MethodDelete, path: "/api/v1/roster-members/m9/link", wantStatus: 404, wantCode: "not_found"},
		{name: "unlinking a member without an approved link", method: http.MethodDelete, path: "/api/v1/roster-members/m2/link", wantStatus: 409, wantCode: "state_conflict"},
		{name: "an invite link for a member of another team", method: http.MethodPost, path: "/api/v1/roster-members/m9/invite-links", wantStatus: 404, wantCode: "not_found"},
		{name: "an invite link for a member with an approved link", method: http.MethodPost, path: "/api/v1/roster-members/m1/invite-links", wantStatus: 409, wantCode: "state_conflict"},
		{name: "an invite link for a member with a pending link", method: http.MethodPost, path: "/api/v1/roster-members/m2/invite-links", wantStatus: 409, wantCode: "state_conflict"},
		{name: "an invite link for a member who has left", method: http.MethodPost, path: "/api/v1/roster-members/m3/invite-links", wantStatus: 409, wantCode: "state_conflict"},
		{name: "approving an approved link", method: http.MethodPost, path: "/api/v1/member-links/k1/approve", wantStatus: 409, wantCode: "state_conflict"},
		{name: "rejecting an approved link", method: http.MethodPost, path: "/api/v1/member-links/k1/reject", wantStatus: 409, wantCode: "state_conflict"},
		{name: "approving an unknown link", method: http.MethodPost, path: "/api/v1/member-links/k0/approve", wantStatus: 404, wantCode: "not_found"},
		{name: "approving a link of a member of another team", method: http.MethodPost, path: "/api/v1/member-links/k9/approve", wantStatus: 404, wantCode: "not_found"},
		{name: "rejecting a link of a member of another team", method: http.MethodPost, path: "/api/v1/member-links/k9/reject", wantStatus: 404, wantCode: "not_found"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			store := newRosterStore().withLink("k1", "m1", "u1", true).withLink("k2", "m2", "u2", false).withLink("k9", "m9", "u9", false).withLeft("m3")
			rec := serve(newRosterRouter(store), testCase.method, testCase.path, testCase.body, sessionOf(organizerToken))
			if rec.Code != testCase.wantStatus || problemCode(t, rec.Body.Bytes()) != testCase.wantCode {
				t.Errorf("status %d, body %s, want %d %s", rec.Code, rec.Body, testCase.wantStatus, testCase.wantCode)
			}
		})
	}
}
