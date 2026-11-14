package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type runningUnitOfWork struct{ runs int }

func (unitOfWork *runningUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	unitOfWork.runs++
	return work(ctx)
}

type fakeMemberLinkRepository struct {
	byUser map[shared.UserID][]*identity.MemberLink
	saved  []*identity.MemberLink
}

func (fake *fakeMemberLinkRepository) FindByID(context.Context, shared.MemberLinkID) (*identity.MemberLink, error) {
	return nil, shared.ErrNotFound
}

func (fake *fakeMemberLinkRepository) ListByUser(_ context.Context, userID shared.UserID) ([]*identity.MemberLink, error) {
	return fake.byUser[userID], nil
}

func (fake *fakeMemberLinkRepository) ListByMember(context.Context, shared.MemberID) ([]*identity.MemberLink, error) {
	return nil, nil
}

func (fake *fakeMemberLinkRepository) Save(_ context.Context, link *identity.MemberLink) error {
	fake.saved = append(fake.saved, link)
	return nil
}

type fakeRosterMemberRepository struct {
	byID map[shared.MemberID]*roster.Member
}

func (fake fakeRosterMemberRepository) FindByID(_ context.Context, id shared.MemberID) (*roster.Member, error) {
	if member, ok := fake.byID[id]; ok {
		return member, nil
	}
	return nil, shared.ErrNotFound
}

func (fake fakeRosterMemberRepository) Save(context.Context, *roster.Member) error { return nil }

// completeLineLoginFixture is the state the callback runs against: one login attempt
// started with the state "state", and the roster members m1 (active) and m2 (left).
type completeLineLoginFixture struct {
	unitOfWork  *runningUnitOfWork
	attempts    *fakeLoginAttemptRepository
	lineLogin   *fakeLineLogin
	users       *fakeUserRepository
	memberLinks *fakeMemberLinkRepository
	inviteQuery fakeInviteLinkQueryService
	inviteLinks *fakeInviteLinkRepository
	members     fakeRosterMemberRepository
	codes       *fakeLoginCodeRepository
	at          func() time.Time
}

// newCompleteLineLoginFixture starts the login from invite, or without one when invite is nil.
func newCompleteLineLoginFixture(t *testing.T, invite *identity.InviteLink) *completeLineLoginFixture {
	t.Helper()
	attempt, _, err := identity.NewLoginAttemptFactory(&sequentialIDs{}, fixedToken("state")).New(invite, now)
	if err != nil {
		t.Fatal(err)
	}
	memberFactory := roster.NewMemberFactory(&sequentialIDs{})
	fixture := &completeLineLoginFixture{
		unitOfWork:  &runningUnitOfWork{},
		attempts:    &fakeLoginAttemptRepository{byHash: map[string]*identity.LoginAttempt{string(attempt.StateHash()): attempt}},
		lineLogin:   &fakeLineLogin{identity: LineIdentity{UserID: "U-line", DisplayName: "Taro"}},
		users:       &fakeUserRepository{byLineUserID: map[string]*identity.User{}},
		memberLinks: &fakeMemberLinkRepository{byUser: map[shared.UserID][]*identity.MemberLink{}},
		inviteQuery: fakeInviteLinkQueryService{byID: map[shared.InviteLinkID]InviteLinkRecord{}},
		inviteLinks: &fakeInviteLinkRepository{},
		members: fakeRosterMemberRepository{byID: map[shared.MemberID]*roster.Member{
			"m1": memberFactory.Reconstruct("m1", "t1", "Taro Yamada", nil, roster.DefaultAverage(), 0),
			"m2": memberFactory.Reconstruct("m2", "t1", "Hanako", timeAt(-1), roster.DefaultAverage(), 0),
		}},
		codes: &fakeLoginCodeRepository{},
		at:    clock,
	}
	if invite != nil {
		fixture.inviteQuery.byID[invite.ID()] = inviteRecord(invite)
	}
	return fixture
}

func (fixture *completeLineLoginFixture) useCase() *CompleteLineLoginUseCase {
	return NewCompleteLineLoginUseCase(fixture.unitOfWork, fixture.attempts, fixture.lineLogin,
		fixture.users, identity.NewUserFactory(&sequentialIDs{}),
		fixture.memberLinks, identity.NewMemberLinkFactory(&sequentialIDs{}),
		fixture.inviteQuery, fixture.inviteLinks, identity.NewInviteLinkFactory(&sequentialIDs{}, fixedToken("unused")),
		fixture.members, fixture.codes, identity.NewLoginCodeFactory(&sequentialIDs{}, fixedToken("login-code")), fixture.at)
}

func (fixture *completeLineLoginFixture) execute(code, lineError string) (string, error) {
	return fixture.useCase().Execute(context.Background(), CompleteLineLoginCommandFactory{}.Create("state", code, lineError))
}

// existingUser saves a user of the LINE user U-line with the given links.
func (fixture *completeLineLoginFixture) existingUser(links ...*identity.MemberLink) {
	user := identity.NewUserFactory(&sequentialIDs{}).Reconstruct("u1", "U-line", "Old name")
	fixture.users.byLineUserID["U-line"] = user
	fixture.memberLinks.byUser["u1"] = links
}

func inviteLinkOf(t *testing.T, memberID shared.MemberID) *identity.InviteLink {
	t.Helper()
	link, _, err := identity.NewInviteLinkFactory(&sequentialIDs{}, fixedToken("invite-token")).New(memberID, now)
	if err != nil {
		t.Fatal(err)
	}
	return link
}

func memberLink(memberID shared.MemberID, approvedAt *time.Time) *identity.MemberLink {
	return identity.NewMemberLinkFactory(&sequentialIDs{}).Reconstruct("l1", "u1", memberID, "i0", now.Add(-time.Hour), approvedAt, nil, nil)
}

func TestCompleteLineLoginUseCase(t *testing.T) {
	t.Run("a first login without an invite link creates the user and a code", func(t *testing.T) {
		fixture := newCompleteLineLoginFixture(t, nil)
		got, err := fixture.execute("line-code", "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "login-code" {
			t.Errorf("Execute() = %q, want login-code", got)
		}
		if fixture.lineLogin.gotCode != "line-code" || fixture.lineLogin.gotVerifier != "state" || fixture.lineLogin.gotNonce != "state" {
			t.Error("LINE was not given the code with the attempt's verifier and nonce")
		}
		if len(fixture.users.saved) != 1 || fixture.users.saved[0].LineUserID() != "U-line" || fixture.users.saved[0].DisplayName() != "Taro" {
			t.Fatal("the user of the ID token was not saved")
		}
		if len(fixture.codes.saved) != 1 || fixture.codes.saved[0].UserID() != fixture.users.saved[0].ID() {
			t.Error("the code was not saved for the user")
		}
		if len(fixture.memberLinks.saved) != 0 || fixture.unitOfWork.runs != 1 {
			t.Errorf("saved %d member links in %d units of work, want 0 in 1", len(fixture.memberLinks.saved), fixture.unitOfWork.runs)
		}
	})

	t.Run("a returning user gets the display name of LINE", func(t *testing.T) {
		fixture := newCompleteLineLoginFixture(t, nil)
		fixture.existingUser()
		if _, err := fixture.execute("line-code", ""); err != nil {
			t.Fatal(err)
		}
		if len(fixture.users.saved) != 1 || fixture.users.saved[0].ID() != "u1" || fixture.users.saved[0].DisplayName() != "Taro" {
			t.Error("the user u1 was not saved with the new display name")
		}
	})

	t.Run("an invite link creates a pending link and is used up", func(t *testing.T) {
		invite := inviteLinkOf(t, "m1")
		fixture := newCompleteLineLoginFixture(t, invite)
		if _, err := fixture.execute("line-code", ""); err != nil {
			t.Fatal(err)
		}
		if len(fixture.memberLinks.saved) != 1 {
			t.Fatalf("saved %d member links, want 1", len(fixture.memberLinks.saved))
		}
		link := fixture.memberLinks.saved[0]
		if !link.Pending() || link.MemberID() != "m1" || link.InviteLinkID() != invite.ID() || link.UserID() != fixture.users.saved[0].ID() {
			t.Error("the link is not a pending link of the user to m1 through the invite link")
		}
		if len(fixture.inviteLinks.saved) != 1 {
			t.Fatalf("saved %d invite links, want 1", len(fixture.inviteLinks.saved))
		}
		if _, used := fixture.inviteLinks.saved[0].UsedAt(); !used {
			t.Error("the invite link was not used up")
		}
	})

	for _, existing := range []struct {
		name string
		link *identity.MemberLink
	}{
		{name: "pending", link: memberLink("m1", nil)},
		{name: "approved", link: memberLink("m1", timeAt(-1))},
	} {
		t.Run("a user with a "+existing.name+" link gets no new link from an invite link", func(t *testing.T) {
			fixture := newCompleteLineLoginFixture(t, inviteLinkOf(t, "m1"))
			fixture.existingUser(existing.link)
			got, err := fixture.execute("line-code", "")
			if err != nil || got != "login-code" {
				t.Fatalf("Execute() = %q, %v, want login-code", got, err)
			}
			if len(fixture.memberLinks.saved) != 0 || len(fixture.inviteLinks.saved) != 0 {
				t.Error("a member link or the invite link was saved")
			}
		})
	}

	usedInvite := inviteLinkOf(t, "m1")
	usedFixture := newCompleteLineLoginFixture(t, usedInvite)
	usedRecord := inviteRecord(usedInvite)
	usedRecord.UsedAt = timeAt(0)
	usedFixture.inviteQuery.byID[usedInvite.ID()] = usedRecord

	approvedLeft := newCompleteLineLoginFixture(t, nil)
	approvedLeft.existingUser(memberLink("m2", timeAt(-1)))

	expired := newCompleteLineLoginFixture(t, nil)
	expired.at = func() time.Time { return now.Add(identity.LoginAttemptLifetime) }

	lineFailed := newCompleteLineLoginFixture(t, nil)
	lineFailed.lineLogin.err = errors.New("invalid ID token")

	tests := []struct {
		name      string
		fixture   *completeLineLoginFixture
		state     string
		code      string
		lineError string
		want      error
	}{
		{name: "the invite link's member has left", fixture: newCompleteLineLoginFixture(t, inviteLinkOf(t, "m2")), code: "line-code", want: ErrMemberLeft},
		{name: "the member of the approved link has left", fixture: approvedLeft, code: "line-code", want: ErrMemberLeft},
		{name: "the invite link was used after the login started", fixture: usedFixture, code: "line-code", want: identity.ErrInviteLinkUnusable},
		{name: "unknown state", fixture: newCompleteLineLoginFixture(t, nil), state: "unknown", code: "line-code", want: ErrLoginFailed},
		{name: "LINE set an error", fixture: newCompleteLineLoginFixture(t, nil), lineError: "access_denied", want: ErrLoginFailed},
		{name: "no code", fixture: newCompleteLineLoginFixture(t, nil), want: ErrLoginFailed},
		{name: "the callback came after 10 minutes", fixture: expired, code: "line-code", want: ErrLoginFailed},
		{name: "the ID token is not valid", fixture: lineFailed, code: "line-code", want: ErrLoginFailed},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			state := testCase.state
			if state == "" {
				state = "state"
			}
			_, err := testCase.fixture.useCase().Execute(context.Background(), CompleteLineLoginCommandFactory{}.Create(state, testCase.code, testCase.lineError))
			if !errors.Is(err, testCase.want) {
				t.Errorf("Execute() error = %v, want %v", err, testCase.want)
			}
			if len(testCase.fixture.codes.saved) != 0 {
				t.Error("a code was saved")
			}
		})
	}

	t.Run("a state can be used once", func(t *testing.T) {
		fixture := newCompleteLineLoginFixture(t, nil)
		if _, err := fixture.execute("line-code", ""); err != nil {
			t.Fatal(err)
		}
		if _, err := fixture.execute("line-code", ""); !errors.Is(err, ErrLoginFailed) {
			t.Errorf("second Execute() error = %v, want ErrLoginFailed", err)
		}
	})
}
