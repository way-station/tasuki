package presentation

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

// lineLoginStore keeps what LINE Login reads and writes in memory, standing in for
// PostgreSQL. The invite link of the token "invite-token" belongs to roster member m1 of the
// team "Tasuki AC"; m2 has left.
type lineLoginStore struct {
	*memoryStore
	invites  map[string]*identity.InviteLink
	attempts map[string]*identity.LoginAttempt
	codes    map[string]*identity.LoginCode
	links    []*identity.MemberLink
	members  map[shared.MemberID]*roster.Member
	// membersErr makes reading roster members fail, as a database that is down does.
	membersErr error
}

func newLineLoginStore(t *testing.T) *lineLoginStore {
	t.Helper()
	invite, _, err := identity.NewInviteLinkFactory(&sequentialIDs{}, fixedToken("invite-token")).New("m1", testNow)
	if err != nil {
		t.Fatal(err)
	}
	leftAt := testNow.Add(-time.Hour)
	memberFactory := roster.NewMemberFactory(&sequentialIDs{})
	return &lineLoginStore{
		memoryStore: newMemoryStore(),
		invites:     map[string]*identity.InviteLink{string(invite.TokenHash()): invite},
		attempts:    map[string]*identity.LoginAttempt{},
		codes:       map[string]*identity.LoginCode{},
		members: map[shared.MemberID]*roster.Member{
			"m1": memberFactory.Reconstruct("m1", "t1", "Taro Yamada", nil, roster.DefaultAverage(), 0),
			"m2": memberFactory.Reconstruct("m2", "t1", "Hanako", &leftAt, roster.DefaultAverage(), 0),
		},
	}
}

// addInvite stores an invite link of the token for the member.
func (store *lineLoginStore) addInvite(t *testing.T, token string, memberID shared.MemberID) {
	t.Helper()
	invite, _, err := identity.NewInviteLinkFactory(&sequentialIDs{count: 100}, fixedToken(token)).New(memberID, testNow)
	if err != nil {
		t.Fatal(err)
	}
	store.invites[string(invite.TokenHash())] = invite
}

func (store *lineLoginStore) inviteRecord(invite *identity.InviteLink) auth.InviteLinkRecord {
	record := auth.InviteLinkRecord{ID: invite.ID(), MemberID: invite.MemberID(), TokenHash: invite.TokenHash(), ExpiresAt: invite.ExpiresAt(), TeamName: "Tasuki AC"}
	if usedAt, ok := invite.UsedAt(); ok {
		record.UsedAt = &usedAt
	}
	if revokedAt, ok := invite.RevokedAt(); ok {
		record.RevokedAt = &revokedAt
	}
	return record
}

type memoryInviteQuery struct{ store *lineLoginStore }

func (query memoryInviteQuery) FindByTokenHash(_ context.Context, hash []byte) (auth.InviteLinkRecord, error) {
	if invite, ok := query.store.invites[string(hash)]; ok {
		return query.store.inviteRecord(invite), nil
	}
	return auth.InviteLinkRecord{}, shared.ErrNotFound
}

func (query memoryInviteQuery) FindByID(_ context.Context, id shared.InviteLinkID) (auth.InviteLinkRecord, error) {
	for _, invite := range query.store.invites {
		if invite.ID() == id {
			return query.store.inviteRecord(invite), nil
		}
	}
	return auth.InviteLinkRecord{}, shared.ErrNotFound
}

type memoryInviteLinks struct{ store *lineLoginStore }

func (repository memoryInviteLinks) FindByTokenHash(_ context.Context, hash []byte) (*identity.InviteLink, error) {
	if invite, ok := repository.store.invites[string(hash)]; ok {
		return invite, nil
	}
	return nil, shared.ErrNotFound
}

func (repository memoryInviteLinks) FindOpenByMember(context.Context, shared.MemberID) (*identity.InviteLink, error) {
	return nil, shared.ErrNotFound
}

func (repository memoryInviteLinks) Save(_ context.Context, invite *identity.InviteLink) error {
	repository.store.invites[string(invite.TokenHash())] = invite
	return nil
}

type memoryAttempts struct{ store *lineLoginStore }

func (repository memoryAttempts) Save(_ context.Context, attempt *identity.LoginAttempt) error {
	repository.store.attempts[string(attempt.StateHash())] = attempt
	return nil
}

func (repository memoryAttempts) Take(_ context.Context, stateHash []byte) (*identity.LoginAttempt, error) {
	attempt, ok := repository.store.attempts[string(stateHash)]
	if !ok {
		return nil, shared.ErrNotFound
	}
	delete(repository.store.attempts, string(stateHash))
	return attempt, nil
}

type memoryCodes struct{ store *lineLoginStore }

func (repository memoryCodes) Save(_ context.Context, code *identity.LoginCode) error {
	repository.store.codes[string(code.CodeHash())] = code
	return nil
}

func (repository memoryCodes) Take(_ context.Context, codeHash []byte) (*identity.LoginCode, error) {
	code, ok := repository.store.codes[string(codeHash)]
	if !ok {
		return nil, shared.ErrNotFound
	}
	delete(repository.store.codes, string(codeHash))
	return code, nil
}

type memoryMemberLinks struct{ store *lineLoginStore }

func (repository memoryMemberLinks) FindByID(context.Context, shared.MemberLinkID) (*identity.MemberLink, error) {
	return nil, shared.ErrNotFound
}

func (repository memoryMemberLinks) ListByUser(_ context.Context, userID shared.UserID) ([]*identity.MemberLink, error) {
	var links []*identity.MemberLink
	for _, link := range repository.store.links {
		if link.UserID() == userID {
			links = append(links, link)
		}
	}
	return links, nil
}

func (repository memoryMemberLinks) ListByMember(context.Context, shared.MemberID) ([]*identity.MemberLink, error) {
	return nil, nil
}

func (repository memoryMemberLinks) Save(_ context.Context, link *identity.MemberLink) error {
	repository.store.links = append(repository.store.links, link)
	return nil
}

type memoryRosterMembers struct{ store *lineLoginStore }

func (repository memoryRosterMembers) FindByID(_ context.Context, id shared.MemberID) (*roster.Member, error) {
	if repository.store.membersErr != nil {
		return nil, repository.store.membersErr
	}
	if member, ok := repository.store.members[id]; ok {
		return member, nil
	}
	return nil, shared.ErrNotFound
}

func (repository memoryRosterMembers) Save(context.Context, *roster.Member) error { return nil }

// fakeLINE stands in for LINE: the authorization code "line-code" names the LINE user
// U-line, and any other code fails as an ID token that is not valid.
type fakeLINE struct{}

func (fakeLINE) AuthorizationURL(state, _, _ string) string {
	return "https://access.line.me/oauth2/v2.1/authorize?state=" + url.QueryEscape(state)
}

func (fakeLINE) Authenticate(_ context.Context, code, _, _ string) (auth.LineIdentity, error) {
	if code != "line-code" {
		return auth.LineIdentity{}, errors.New("the ID token is not valid")
	}
	return auth.LineIdentity{UserID: "U-line", DisplayName: "Taro"}, nil
}

// newLineLoginRouter serves LINE Login against store; lineLogin registers the routes that
// need LINE. The state of every login is "state-token", and every one-time code is
// "login-code".
func newLineLoginRouter(store *lineLoginStore, lineLogin bool) http.Handler {
	clock := func() time.Time { return testNow }
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sessions := memorySessions{store: store.memoryStore}
	inviteFactory := identity.NewInviteLinkFactory(&sequentialIDs{}, fixedToken("unused"))
	lineLoginController := controller.NewLineLoginController(
		auth.NewCheckInviteLinkUseCase(memoryInviteQuery{store: store}, inviteFactory, auth.InviteLinkDTOFactory{}, clock),
		auth.NewStartLineLoginUseCase(memoryInviteLinks{store: store}, memoryAttempts{store: store}, identity.NewLoginAttemptFactory(&sequentialIDs{}, fixedToken("state-token")), fakeLINE{}, clock),
		auth.NewCompleteLineLoginUseCase(passUnitOfWork{}, memoryAttempts{store: store}, fakeLINE{},
			memoryUsers{store: store.memoryStore}, identity.NewUserFactory(&sequentialIDs{}),
			memoryMemberLinks{store: store}, identity.NewMemberLinkFactory(&sequentialIDs{}),
			memoryInviteQuery{store: store}, memoryInviteLinks{store: store}, inviteFactory,
			memoryRosterMembers{store: store}, memoryCodes{store: store}, identity.NewLoginCodeFactory(&sequentialIDs{}, fixedToken("login-code")), clock),
		auth.NewExchangeLoginCodeUseCase(memoryCodes{store: store}, sessions, identity.NewSessionFactory(&sequentialIDs{}, fixedToken("session-token")), auth.SessionDTOFactory{}, clock),
		auth.StartLineLoginCommandFactory{},
		auth.CompleteLineLoginCommandFactory{},
		testOrigin,
		logger,
	)
	check := health.NewCheckUseCase(fakeHealth{}, fakeHealth{}, health.StatusDTOFactory{}, logger)
	return NewRouter(Controllers{Health: controller.NewHealthController(check), LineLogin: lineLoginController}, Options{
		Authenticate:   middleware.Authenticate(auth.NewAuthenticateUseCase(sessions, clock), logger),
		ResolveCaller:  middleware.ResolveCaller(tokenCallers{}, logger),
		FrontendOrigin: testOrigin,
		LineLogin:      lineLogin,
		Logger:         logger,
	})
}

// callback sends LINE's redirect with the query and returns the query of the frontend URL
// it redirects to.
func callback(t *testing.T, handler http.Handler, query string) url.Values {
	t.Helper()
	rec := serveFrom(handler, "", http.MethodGet, "/api/v1/auth/line/callback?"+query, "", nil)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302: %s", rec.Code, rec.Body)
	}
	location, err := url.Parse(rec.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if got := location.Scheme + "://" + location.Host + location.Path; got != testOrigin+"/auth/complete" {
		t.Fatalf("Location = %s, want %s/auth/complete", location, testOrigin)
	}
	return location.Query()
}

func authorize(t *testing.T, handler http.Handler, body string) {
	t.Helper()
	if rec := serve(handler, http.MethodPost, "/api/v1/auth/line/authorize", body, nil); rec.Code != http.StatusOK {
		t.Fatalf("authorize status = %d: %s", rec.Code, rec.Body)
	}
}

func TestGetInviteLink(t *testing.T) {
	store := newLineLoginStore(t)
	handler := newLineLoginRouter(store, false)

	t.Run("a usable link returns the team name without a session", func(t *testing.T) {
		rec := serve(handler, http.MethodGet, "/api/v1/invite-links/invite-token", "", nil)
		if rec.Code != http.StatusOK || rec.Body.String() != `{"team_name":"Tasuki AC"}`+"\n" {
			t.Errorf("status = %d, body = %s", rec.Code, rec.Body)
		}
	})

	t.Run("an unknown token is 404 invite_unusable", func(t *testing.T) {
		rec := serve(handler, http.MethodGet, "/api/v1/invite-links/unknown", "", nil)
		if rec.Code != http.StatusNotFound || problemCode(t, rec.Body.Bytes()) != "invite_unusable" {
			t.Errorf("status = %d, body = %s", rec.Code, rec.Body)
		}
	})
}

func TestAuthorizeLine(t *testing.T) {
	t.Run("returns LINE's login page without a session", func(t *testing.T) {
		rec := serve(newLineLoginRouter(newLineLoginStore(t), true), http.MethodPost, "/api/v1/auth/line/authorize", `{}`, nil)
		if want := `{"authorization_url":"https://access.line.me/oauth2/v2.1/authorize?state=state-token"}` + "\n"; rec.Code != http.StatusOK || rec.Body.String() != want {
			t.Errorf("status = %d, body = %s, want %s", rec.Code, rec.Body, want)
		}
	})

	tests := []struct {
		name   string
		origin string
		body   string
		want   int
		code   string
	}{
		{name: "an unknown invite token", origin: testOrigin, body: `{"invite_token":"unknown"}`, want: http.StatusNotFound, code: "invite_unusable"},
		{name: "a malformed body", origin: testOrigin, body: `{"invite_token":`, want: http.StatusBadRequest, code: "invalid_request"},
		{name: "a foreign Origin", origin: "https://evil.example", body: `{}`, want: http.StatusForbidden, code: "forbidden"},
		{name: "no Origin", body: `{}`, want: http.StatusForbidden, code: "forbidden"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := serveFrom(newLineLoginRouter(newLineLoginStore(t), true), testCase.origin, http.MethodPost, "/api/v1/auth/line/authorize", testCase.body, nil)
			if rec.Code != testCase.want || problemCode(t, rec.Body.Bytes()) != testCase.code {
				t.Errorf("status = %d, body = %s, want %d %s", rec.Code, rec.Body, testCase.want, testCase.code)
			}
		})
	}

	t.Run("LINE Login routes do not exist while LINE is not configured", func(t *testing.T) {
		handler := newLineLoginRouter(newLineLoginStore(t), false)
		for _, rec := range []*http.Response{
			serve(handler, http.MethodPost, "/api/v1/auth/line/authorize", `{}`, nil).Result(),
			serveFrom(handler, "", http.MethodGet, "/api/v1/auth/line/callback?state=state-token&code=line-code", "", nil).Result(),
		} {
			if rec.StatusCode != http.StatusNotFound && rec.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want the route to be missing", rec.StatusCode)
			}
		}
	})
}

func TestLineCallback(t *testing.T) {
	t.Run("an invite link creates a pending link and redirects with the one-time code", func(t *testing.T) {
		store := newLineLoginStore(t)
		handler := newLineLoginRouter(store, true)
		authorize(t, handler, `{"invite_token":"invite-token"}`)
		query := callback(t, handler, "state=state-token&code=line-code")
		if query.Get("code") != "login-code" || query.Has("error") {
			t.Errorf("query = %v, want code=login-code", query)
		}
		if len(store.links) != 1 || !store.links[0].Pending() || store.links[0].MemberID() != "m1" {
			t.Errorf("links = %v, want one pending link to m1", store.links)
		}
	})

	t.Run("a login without an invite link redirects with the one-time code", func(t *testing.T) {
		handler := newLineLoginRouter(newLineLoginStore(t), true)
		authorize(t, handler, `{}`)
		if query := callback(t, handler, "state=state-token&code=line-code"); query.Get("code") != "login-code" {
			t.Errorf("query = %v, want code=login-code", query)
		}
	})

	t.Run("the member of the invite link has left", func(t *testing.T) {
		store := newLineLoginStore(t)
		store.addInvite(t, "left-token", "m2")
		handler := newLineLoginRouter(store, true)
		authorize(t, handler, `{"invite_token":"left-token"}`)
		if query := callback(t, handler, "state=state-token&code=line-code"); query.Get("error") != "left" || query.Has("code") {
			t.Errorf("query = %v, want error=left", query)
		}
	})

	t.Run("the invite link became unusable after the login started", func(t *testing.T) {
		store := newLineLoginStore(t)
		handler := newLineLoginRouter(store, true)
		authorize(t, handler, `{"invite_token":"invite-token"}`)
		for _, invite := range store.invites {
			invite.Revoke(testNow)
		}
		if query := callback(t, handler, "state=state-token&code=line-code"); query.Get("error") != "invite_unusable" {
			t.Errorf("query = %v, want error=invite_unusable", query)
		}
	})

	failures := []struct {
		name  string
		query string
	}{
		{name: "LINE set an error", query: "state=state-token&error=access_denied"},
		{name: "the ID token is not valid", query: "state=state-token&code=other"},
		{name: "an unknown state", query: "state=unknown&code=line-code"},
	}
	for _, testCase := range failures {
		t.Run(testCase.name, func(t *testing.T) {
			handler := newLineLoginRouter(newLineLoginStore(t), true)
			authorize(t, handler, `{}`)
			if query := callback(t, handler, testCase.query); query.Get("error") != "login_failed" || query.Has("code") {
				t.Errorf("query = %v, want error=login_failed", query)
			}
		})
	}

	t.Run("an unexpected failure redirects with login_failed", func(t *testing.T) {
		store := newLineLoginStore(t)
		store.membersErr = errors.New("database is down")
		handler := newLineLoginRouter(store, true)
		authorize(t, handler, `{"invite_token":"invite-token"}`)
		if query := callback(t, handler, "state=state-token&code=line-code"); query.Get("error") != "login_failed" {
			t.Errorf("query = %v, want error=login_failed", query)
		}
	})

	t.Run("a callback without the state is 400", func(t *testing.T) {
		rec := serveFrom(newLineLoginRouter(newLineLoginStore(t), true), "", http.MethodGet, "/api/v1/auth/line/callback?code=line-code", "", nil)
		if rec.Code != http.StatusBadRequest || problemCode(t, rec.Body.Bytes()) != "invalid_request" {
			t.Errorf("status = %d, body = %s", rec.Code, rec.Body)
		}
	})
}

func TestCreateSession(t *testing.T) {
	// loggedIn runs LINE Login up to the callback, so the code "login-code" can be exchanged.
	loggedIn := func(t *testing.T) http.Handler {
		t.Helper()
		handler := newLineLoginRouter(newLineLoginStore(t), true)
		authorize(t, handler, `{}`)
		callback(t, handler, "state=state-token&code=line-code")
		return handler
	}

	t.Run("sets the session cookie and keeps the token out of the body", func(t *testing.T) {
		rec := serve(loggedIn(t), http.MethodPost, "/api/v1/session", `{"code":"login-code"}`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
		}
		cookie := sessionCookieOf(t, rec)
		if cookie.Value != "session-token" || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/api" {
			t.Errorf("cookie = %+v, want the session token as HttpOnly, Secure, SameSite=Lax, Path=/api", cookie)
		}
		if strings.Contains(rec.Body.String(), "session-token") || !strings.Contains(rec.Body.String(), `"expires_at"`) {
			t.Errorf("body = %s, want only expires_at", rec.Body)
		}
	})

	t.Run("a code can be exchanged once", func(t *testing.T) {
		handler := loggedIn(t)
		serve(handler, http.MethodPost, "/api/v1/session", `{"code":"login-code"}`, nil)
		rec := serve(handler, http.MethodPost, "/api/v1/session", `{"code":"login-code"}`, nil)
		if rec.Code != http.StatusBadRequest || problemCode(t, rec.Body.Bytes()) != "invalid_request" || rec.Header().Get("Set-Cookie") != "" {
			t.Errorf("status = %d, Set-Cookie = %q, body = %s", rec.Code, rec.Header().Get("Set-Cookie"), rec.Body)
		}
	})

	tests := []struct {
		name   string
		origin string
		body   string
		want   int
		code   string
	}{
		{name: "an unknown code", origin: testOrigin, body: `{"code":"unknown"}`, want: http.StatusBadRequest, code: "invalid_request"},
		{name: "a malformed body", origin: testOrigin, body: `{"code":`, want: http.StatusBadRequest, code: "invalid_request"},
		{name: "a foreign Origin", origin: "https://evil.example", body: `{"code":"login-code"}`, want: http.StatusForbidden, code: "forbidden"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			rec := serveFrom(loggedIn(t), testCase.origin, http.MethodPost, "/api/v1/session", testCase.body, nil)
			if rec.Code != testCase.want || problemCode(t, rec.Body.Bytes()) != testCase.code || rec.Header().Get("Set-Cookie") != "" {
				t.Errorf("status = %d, Set-Cookie = %q, body = %s, want %d %s", rec.Code, rec.Header().Get("Set-Cookie"), rec.Body, testCase.want, testCase.code)
			}
		})
	}
}
