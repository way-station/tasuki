package rosters

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// teamMember returns a roster member of the team who has not left.
func teamMember(id shared.MemberID, team shared.TeamID) *roster.Member {
	return roster.NewMemberFactory(nil).Reconstruct(id, team, "Taro", nil, roster.DefaultAverage(), 0)
}

type transactionKey struct{}

func inTransaction(ctx context.Context) bool {
	marked, _ := ctx.Value(transactionKey{}).(bool)
	return marked
}

type fakeUnitOfWork struct{}

func (fakeUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	return work(context.WithValue(ctx, transactionKey{}, true))
}

// fakeLinkRepository keeps member links by ID.
type fakeLinkRepository struct {
	stored map[shared.MemberLinkID]*identity.MemberLink
	saved  []*identity.MemberLink
}

func linksOf(links ...*identity.MemberLink) *fakeLinkRepository {
	fake := &fakeLinkRepository{stored: map[shared.MemberLinkID]*identity.MemberLink{}}
	for _, link := range links {
		fake.stored[link.ID()] = link
	}
	return fake
}

func (fake *fakeLinkRepository) FindByID(_ context.Context, id shared.MemberLinkID) (*identity.MemberLink, error) {
	link, ok := fake.stored[id]
	if !ok {
		return nil, fmt.Errorf("member link %s: %w", id, shared.ErrNotFound)
	}
	return link, nil
}

func (fake *fakeLinkRepository) ListByUser(context.Context, shared.UserID) ([]*identity.MemberLink, error) {
	return nil, errors.New("links are not listed by user in this package")
}

func (fake *fakeLinkRepository) ListByMember(_ context.Context, memberID shared.MemberID) ([]*identity.MemberLink, error) {
	var links []*identity.MemberLink
	for _, link := range fake.stored {
		if link.MemberID() == memberID {
			links = append(links, link)
		}
	}
	return links, nil
}

func (fake *fakeLinkRepository) Save(ctx context.Context, link *identity.MemberLink) error {
	if !inTransaction(ctx) {
		return errors.New("the link was saved outside a unit of work")
	}
	fake.stored[link.ID()] = link
	fake.saved = append(fake.saved, link)
	return nil
}

// link returns a link of the member to the user; approved and rejected say which state
// holds.
func link(id shared.MemberLinkID, userID shared.UserID, memberID shared.MemberID, approved, rejected bool) *identity.MemberLink {
	var approvedAt, rejectedAt *time.Time
	if approved {
		approvedAt = at(-time.Hour)
	}
	if rejected {
		rejectedAt = at(-time.Hour)
	}
	return identity.NewMemberLinkFactory(nil).Reconstruct(id, userID, memberID, "i1", testNow.Add(-2*time.Hour), approvedAt, rejectedAt, nil)
}

type fakeSessions struct {
	active map[shared.UserID][]*identity.Session
	saved  []*identity.Session
}

func (fake *fakeSessions) ListActiveByUser(_ context.Context, userID shared.UserID, _ time.Time) ([]*identity.Session, error) {
	return fake.active[userID], nil
}

func (fake *fakeSessions) FindByTokenHash(context.Context, []byte) (*identity.Session, error) {
	return nil, errors.New("sessions are not found by token in this package")
}

func (fake *fakeSessions) Save(ctx context.Context, session *identity.Session) error {
	if !inTransaction(ctx) {
		return errors.New("the session was saved outside a unit of work")
	}
	fake.saved = append(fake.saved, session)
	return nil
}

func session(id shared.SessionID, userID shared.UserID) *identity.Session {
	return identity.NewSessionFactory(nil, nil).Reconstruct(id, userID, nil, testNow.Add(time.Hour), nil)
}

type fakeCanceller struct{ cancelled []shared.RaceID }

func (fake *fakeCanceller) CancelInvitation(ctx context.Context, raceID shared.RaceID, _ shared.MemberID) error {
	if !inTransaction(ctx) {
		return errors.New("the invitation was cancelled outside a unit of work")
	}
	fake.cancelled = append(fake.cancelled, raceID)
	return nil
}

func date(t *testing.T, text string) shared.Date {
	t.Helper()
	parsed, err := shared.ParseDate(text)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

type setLeftFixture struct {
	members  *fakeMemberRepository
	query    *fakeRosterQueryService
	canceled *fakeCanceller
	links    *fakeLinkRepository
	sessions *fakeSessions
}

func (fixture setLeftFixture) useCase(now time.Time) *SetLeftUseCase {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	return NewSetLeftUseCase(fakeUnitOfWork{}, fixture.members, fixture.query, fixture.canceled, fixture.links, fixture.sessions, fixture.sessions, roster.LeaveService{}, func() time.Time { return now }, tokyo, teamID)
}

func newSetLeftFixture(t *testing.T, member *roster.Member) setLeftFixture {
	return setLeftFixture{
		members:  membersOf(member),
		query:    &fakeRosterQueryService{invited: []roster.InvitedRace{{RaceID: "past", RaceDate: date(t, "2026-11-13")}, {RaceID: "today", RaceDate: date(t, "2026-11-14")}, {RaceID: "later", RaceDate: date(t, "2026-12-06")}}},
		canceled: &fakeCanceller{},
		links:    linksOf(link("k1", "u1", "m1", true, false), link("k2", "u2", "m1", false, true)),
		sessions: &fakeSessions{active: map[shared.UserID][]*identity.Session{"u1": {session("s1", "u1"), session("s2", "u1")}, "u2": {session("s3", "u2")}}},
	}
}

func TestSetLeftUseCase(t *testing.T) {
	// 2026-11-13 20:00 UTC is 2026-11-14 05:00 in Japan, so today is 2026-11-14.
	evening := time.Date(2026, 11, 13, 20, 0, 0, 0, time.UTC)

	t.Run("cancels the invitations from today in Japan and revokes the linked user's sessions", func(t *testing.T) {
		fixture := newSetLeftFixture(t, teamMember("m1", teamID))
		if err := fixture.useCase(evening).Execute(context.Background(), RosterCommandFactory{}.SetLeft("m1", true)); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !fixture.members.stored["m1"].Left() {
			t.Error("the member has not left")
		}
		if !slices.Equal(fixture.canceled.cancelled, []shared.RaceID{"today", "later"}) {
			t.Errorf("cancelled = %v, want today and later", fixture.canceled.cancelled)
		}
		var revoked []shared.SessionID
		for _, saved := range fixture.sessions.saved {
			if _, ok := saved.RevokedAt(); ok {
				revoked = append(revoked, saved.ID())
			}
		}
		if !slices.Equal(revoked, []shared.SessionID{"s1", "s2"}) {
			t.Errorf("revoked = %v, want s1 and s2 of the approved user", revoked)
		}
	})

	t.Run("lets a left member rejoin without bringing invitations back", func(t *testing.T) {
		left := teamMember("m1", teamID)
		left.Leave(testNow)
		fixture := newSetLeftFixture(t, left)
		if err := fixture.useCase(evening).Execute(context.Background(), RosterCommandFactory{}.SetLeft("m1", false)); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if fixture.members.stored["m1"].Left() || len(fixture.canceled.cancelled) != 0 || len(fixture.sessions.saved) != 0 {
			t.Errorf("left, cancelled, sessions = %v, %v, %v, want a rejoined member and nothing else", fixture.members.stored["m1"].Left(), fixture.canceled.cancelled, fixture.sessions.saved)
		}
	})

	t.Run("changes nothing when the member has already left", func(t *testing.T) {
		left := teamMember("m1", teamID)
		left.Leave(testNow)
		fixture := newSetLeftFixture(t, left)
		if err := fixture.useCase(evening).Execute(context.Background(), RosterCommandFactory{}.SetLeft("m1", true)); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(fixture.members.saved) != 0 || len(fixture.canceled.cancelled) != 0 {
			t.Errorf("saved, cancelled = %v, %v, want nothing", fixture.members.saved, fixture.canceled.cancelled)
		}
	})

	t.Run("rejects a member of another team", func(t *testing.T) {
		fixture := newSetLeftFixture(t, teamMember("m1", "t2"))
		if err := fixture.useCase(evening).Execute(context.Background(), RosterCommandFactory{}.SetLeft("m1", true)); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute() error = %v, want shared.ErrNotFound", err)
		}
	})
}
