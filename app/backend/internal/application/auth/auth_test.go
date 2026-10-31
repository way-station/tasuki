package auth

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

var now = time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)

func clock() time.Time { return now }

type sequentialIDs struct{ count int }

func (ids *sequentialIDs) NewID() string { ids.count++; return "id" + strconv.Itoa(ids.count) }

type fixedToken string

func (token fixedToken) NewToken() (string, error) { return string(token), nil }

type fakeSessionQueryService struct {
	userID  shared.UserID
	err     error
	gotHash []byte
	gotNow  time.Time
}

func (fake *fakeSessionQueryService) FindActiveUserID(_ context.Context, hash []byte, at time.Time) (shared.UserID, error) {
	fake.gotHash, fake.gotNow = hash, at
	return fake.userID, fake.err
}

func TestAuthenticateUseCase(t *testing.T) {
	t.Run("returns the user of an active session", func(t *testing.T) {
		sessions := &fakeSessionQueryService{userID: "a1"}
		got, err := NewAuthenticateUseCase(sessions, clock).Execute(context.Background(), "token")
		if err != nil || got != "a1" {
			t.Fatalf("Execute() = %q, %v, want a1", got, err)
		}
		if string(sessions.gotHash) != string(identity.HashToken("token")) || !sessions.gotNow.Equal(now) {
			t.Error("the query service was not asked for the token's hash at now")
		}
	})

	tests := []struct {
		name  string
		token string
		err   error
	}{
		{name: "no token", token: ""},
		{name: "unknown, expired, or revoked session", token: "token", err: shared.ErrNotFound},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			sessions := &fakeSessionQueryService{err: testCase.err}
			_, err := NewAuthenticateUseCase(sessions, clock).Execute(context.Background(), testCase.token)
			if !errors.Is(err, ErrUnauthenticated) {
				t.Errorf("Execute() error = %v, want ErrUnauthenticated", err)
			}
		})
	}
}

type fakeMeQueryService struct {
	record MeRecord
	err    error
}

func (fake fakeMeQueryService) FindMe(context.Context, shared.UserID) (MeRecord, error) {
	return fake.record, fake.err
}

func newGetMeUseCase(query MeQueryService) *GetMeUseCase {
	return NewGetMeUseCase(query, identity.NewUserFactory(&sequentialIDs{}), identity.NewMemberLinkFactory(&sequentialIDs{}),
		identity.LinkStatusService{}, identity.NewOrganizerPolicy([]string{"U-organizer"}), MeDTOFactory{})
}

func timeAt(hours int) *time.Time {
	at := now.Add(time.Duration(hours) * time.Hour)
	return &at
}

func TestGetMeUseCase(t *testing.T) {
	approved := MeLinkRecord{ID: "l1", RosterMemberID: "m1", RosterMemberName: "Taro Yamada", InviteLinkID: "i1", CreatedAt: now, ApprovedAt: timeAt(1)}
	pending := MeLinkRecord{ID: "l2", RosterMemberID: "m2", RosterMemberName: "Hanako", InviteLinkID: "i2", CreatedAt: now}
	rejected := MeLinkRecord{ID: "l3", RosterMemberID: "m3", RosterMemberName: "Jiro", InviteLinkID: "i3", CreatedAt: now, RejectedAt: timeAt(1)}
	unlinked := MeLinkRecord{ID: "l4", RosterMemberID: "m1", RosterMemberName: "Taro Yamada", InviteLinkID: "i4", CreatedAt: now, ApprovedAt: timeAt(1), UnlinkedAt: timeAt(2)}

	tests := []struct {
		name          string
		lineUserID    string
		links         []MeLinkRecord
		wantStatus    identity.CallerStatus
		wantOrganizer bool
		wantMember    string
	}{
		{name: "approved member", lineUserID: "U-member", links: []MeLinkRecord{approved}, wantStatus: identity.CallerApproved, wantMember: "Taro Yamada"},
		{name: "pending", lineUserID: "U-member", links: []MeLinkRecord{pending}, wantStatus: identity.CallerPending},
		{name: "rejected", lineUserID: "U-member", links: []MeLinkRecord{rejected}, wantStatus: identity.CallerRejected},
		{name: "unlinked", lineUserID: "U-member", links: []MeLinkRecord{unlinked}, wantStatus: identity.CallerNone},
		{name: "no link", lineUserID: "U-member", wantStatus: identity.CallerNone},
		{name: "organizer without a link", lineUserID: "U-organizer", wantStatus: identity.CallerNone, wantOrganizer: true},
		{name: "organizer with a link", lineUserID: "U-organizer", links: []MeLinkRecord{approved}, wantStatus: identity.CallerApproved, wantOrganizer: true, wantMember: "Taro Yamada"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query := fakeMeQueryService{record: MeRecord{LineUserID: testCase.lineUserID, DisplayName: "Taro", Links: testCase.links}}
			got, err := newGetMeUseCase(query).Execute(context.Background(), "a1")
			if err != nil {
				t.Fatal(err)
			}
			if got.DisplayName() != "Taro" || got.Status() != testCase.wantStatus || got.IsOrganizer() != testCase.wantOrganizer {
				t.Errorf("got %q, %s, organizer %v; want %s, organizer %v", got.DisplayName(), got.Status(), got.IsOrganizer(), testCase.wantStatus, testCase.wantOrganizer)
			}
			member, ok := got.RosterMember()
			if member.Name() != testCase.wantMember || ok != (testCase.wantMember != "") {
				t.Errorf("RosterMember() = %q, %v, want %q", member.Name(), ok, testCase.wantMember)
			}
		})
	}

	t.Run("a user that no longer exists is unauthenticated", func(t *testing.T) {
		_, err := newGetMeUseCase(fakeMeQueryService{err: shared.ErrNotFound}).Execute(context.Background(), "a1")
		if !errors.Is(err, ErrUnauthenticated) {
			t.Errorf("Execute() error = %v, want ErrUnauthenticated", err)
		}
	})
}

type fakeUserRepository struct {
	byLineUserID map[string]*identity.User
	saved        []*identity.User
}

func (fake *fakeUserRepository) FindByID(context.Context, shared.UserID) (*identity.User, error) {
	return nil, shared.ErrNotFound
}

func (fake *fakeUserRepository) FindByLineUserID(_ context.Context, lineUserID string) (*identity.User, error) {
	if user, ok := fake.byLineUserID[lineUserID]; ok {
		return user, nil
	}
	return nil, shared.ErrNotFound
}

func (fake *fakeUserRepository) Save(_ context.Context, user *identity.User) error {
	fake.saved = append(fake.saved, user)
	return nil
}

type fakeSessionRepository struct {
	byHash map[string]*identity.Session
	saved  []*identity.Session
}

func (fake *fakeSessionRepository) FindByTokenHash(_ context.Context, hash []byte) (*identity.Session, error) {
	if session, ok := fake.byHash[string(hash)]; ok {
		return session, nil
	}
	return nil, shared.ErrNotFound
}

func (fake *fakeSessionRepository) Save(_ context.Context, session *identity.Session) error {
	fake.saved = append(fake.saved, session)
	return nil
}

func newLoginUseCase(users *fakeUserRepository, sessions *fakeSessionRepository) *LoginUseCase {
	return NewLoginUseCase(users, sessions, identity.NewUserFactory(&sequentialIDs{}),
		identity.NewSessionFactory(&sequentialIDs{}, fixedToken("session-token")), SessionDTOFactory{}, clock)
}

func TestLoginUseCase(t *testing.T) {
	command, err := LoginCommandFactory{}.Create(" U-1 ", " Taro ")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("creates the user on the first login, and a session", func(t *testing.T) {
		users := &fakeUserRepository{}
		sessions := &fakeSessionRepository{}

		got, err := newLoginUseCase(users, sessions).Execute(context.Background(), command)
		if err != nil {
			t.Fatal(err)
		}
		if len(users.saved) != 1 || users.saved[0].LineUserID() != "U-1" || users.saved[0].DisplayName() != "Taro" {
			t.Fatalf("saved users = %v, want one for U-1 named Taro", users.saved)
		}
		if len(sessions.saved) != 1 || sessions.saved[0].UserID() != users.saved[0].ID() {
			t.Fatalf("saved sessions = %v, want one for the user", sessions.saved)
		}
		if got.Token() != "session-token" || !got.ExpiresAt().Equal(now.Add(identity.SessionLifetime)) {
			t.Errorf("Execute() = %q, %v", got.Token(), got.ExpiresAt())
		}
	})

	t.Run("keeps the user and updates its display name", func(t *testing.T) {
		existing := identity.NewUserFactory(&sequentialIDs{}).Reconstruct("a1", "U-1", "Old name")
		users := &fakeUserRepository{byLineUserID: map[string]*identity.User{"U-1": existing}}
		sessions := &fakeSessionRepository{}

		if _, err := newLoginUseCase(users, sessions).Execute(context.Background(), command); err != nil {
			t.Fatal(err)
		}
		if len(users.saved) != 1 || users.saved[0].ID() != "a1" || users.saved[0].DisplayName() != "Taro" {
			t.Errorf("saved users = %v, want a1 renamed to Taro", users.saved)
		}
		if sessions.saved[0].UserID() != "a1" {
			t.Errorf("session user = %q, want a1", sessions.saved[0].UserID())
		}
	})
}

func TestLoginCommandFactory(t *testing.T) {
	tests := []struct {
		name        string
		lineUserID  string
		displayName string
		wantField   string
	}{
		{name: "blank LINE user ID", lineUserID: " ", displayName: "Taro", wantField: "line_user_id"},
		{name: "blank display name", lineUserID: "U-1", displayName: "", wantField: "display_name"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := LoginCommandFactory{}.Create(testCase.lineUserID, testCase.displayName)
			var validationErr *shared.ValidationError
			if !errors.As(err, &validationErr) || validationErr.Field != testCase.wantField {
				t.Errorf("Create() error = %v, want a ValidationError for %s", err, testCase.wantField)
			}
		})
	}
}

func TestLogoutUseCase(t *testing.T) {
	t.Run("revokes the session", func(t *testing.T) {
		session, token, err := identity.NewSessionFactory(&sequentialIDs{}, fixedToken("token")).New("a1", now.Add(-time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		sessions := &fakeSessionRepository{byHash: map[string]*identity.Session{string(session.TokenHash()): session}}

		if err := NewLogoutUseCase(sessions, clock).Execute(context.Background(), token); err != nil {
			t.Fatal(err)
		}
		if at, ok := sessions.saved[0].RevokedAt(); !ok || !at.Equal(now) {
			t.Errorf("RevokedAt() = %v, %v, want %v", at, ok, now)
		}
	})

	t.Run("a session that is already gone is not an error", func(t *testing.T) {
		if err := NewLogoutUseCase(&fakeSessionRepository{}, clock).Execute(context.Background(), "unknown"); err != nil {
			t.Errorf("Execute() = %v, want nil", err)
		}
	})
}
