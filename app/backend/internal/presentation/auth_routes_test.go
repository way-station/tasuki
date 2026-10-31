package presentation

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

var testNow = time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)

type sequentialIDs struct{ count int }

func (ids *sequentialIDs) NewID() string { ids.count++; return "id" + strconv.Itoa(ids.count) }

type fixedToken string

func (token fixedToken) NewToken() (string, error) { return string(token), nil }

// memoryStore keeps users and sessions in memory, standing in for PostgreSQL.
type memoryStore struct {
	users    map[string]*identity.User
	sessions map[string]*identity.Session
}

func newMemoryStore() *memoryStore {
	return &memoryStore{users: map[string]*identity.User{}, sessions: map[string]*identity.Session{}}
}

type memoryUsers struct{ store *memoryStore }

func (repository memoryUsers) FindByID(_ context.Context, id shared.UserID) (*identity.User, error) {
	for _, user := range repository.store.users {
		if user.ID() == id {
			return user, nil
		}
	}
	return nil, shared.ErrNotFound
}

func (repository memoryUsers) FindByLineUserID(_ context.Context, lineUserID string) (*identity.User, error) {
	if user, ok := repository.store.users[lineUserID]; ok {
		return user, nil
	}
	return nil, shared.ErrNotFound
}

func (repository memoryUsers) Save(_ context.Context, user *identity.User) error {
	repository.store.users[user.LineUserID()] = user
	return nil
}

type memorySessions struct{ store *memoryStore }

func (repository memorySessions) FindByTokenHash(_ context.Context, hash []byte) (*identity.Session, error) {
	if session, ok := repository.store.sessions[string(hash)]; ok {
		return session, nil
	}
	return nil, shared.ErrNotFound
}

func (repository memorySessions) Save(_ context.Context, session *identity.Session) error {
	repository.store.sessions[string(session.TokenHash())] = session
	return nil
}

func (repository memorySessions) FindActiveUserID(_ context.Context, hash []byte, now time.Time) (shared.UserID, error) {
	if session, ok := repository.store.sessions[string(hash)]; ok && session.Active(now) {
		return session.UserID(), nil
	}
	return "", shared.ErrNotFound
}

type memoryMe struct{ store *memoryStore }

func (query memoryMe) FindMe(ctx context.Context, userID shared.UserID) (auth.MeRecord, error) {
	user, err := memoryUsers(query).FindByID(ctx, userID)
	if err != nil {
		return auth.MeRecord{}, err
	}
	return auth.MeRecord{LineUserID: user.LineUserID(), DisplayName: user.DisplayName()}, nil
}

type fakeHealth struct{}

func (fakeHealth) PingMongo(context.Context) error    { return nil }
func (fakeHealth) PingPostgres(context.Context) error { return nil }

func newAuthRouter(devLogin bool) http.Handler {
	store := newMemoryStore()
	clock := func() time.Time { return testNow }
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userFactory := identity.NewUserFactory(&sequentialIDs{})
	sessionFactory := identity.NewSessionFactory(&sequentialIDs{}, fixedToken("session-token"))
	sessions := memorySessions{store: store}

	authenticate := auth.NewAuthenticateUseCase(sessions, clock)
	authController := controller.NewAuthController(
		auth.NewLoginUseCase(memoryUsers{store: store}, sessions, userFactory, sessionFactory, auth.SessionDTOFactory{}, clock),
		auth.NewLogoutUseCase(sessions, clock),
		auth.NewGetMeUseCase(memoryMe{store: store}, userFactory, identity.NewMemberLinkFactory(&sequentialIDs{}),
			identity.LinkStatusService{}, identity.NewOrganizerPolicy([]string{"U-organizer"}), auth.MeDTOFactory{}),
		auth.LoginCommandFactory{},
	)
	check := health.NewCheckUseCase(fakeHealth{}, fakeHealth{}, health.StatusDTOFactory{}, logger)
	return NewRouter(Controllers{Health: controller.NewHealthController(check), Auth: authController}, Options{
		Authenticate: middleware.Authenticate(authenticate, logger),
		DevLogin:     devLogin,
		Logger:       logger,
	})
}

func serve(handler http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func sessionCookieOf(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == middleware.SessionCookieName {
			return cookie
		}
	}
	t.Fatalf("no %s cookie in %v", middleware.SessionCookieName, rec.Header())
	return nil
}

func TestDevLogin(t *testing.T) {
	t.Run("does not exist unless enabled (ADR-0009)", func(t *testing.T) {
		rec := serve(newAuthRouter(false), http.MethodPost, "/api/v1/dev/login", `{"line_user_id":"U-1","display_name":"Taro"}`, nil)
		if rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want the route to be missing", rec.Code)
		}
	})

	t.Run("sets the session cookie", func(t *testing.T) {
		rec := serve(newAuthRouter(true), http.MethodPost, "/api/v1/dev/login", `{"line_user_id":"U-1","display_name":"Taro"}`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
		}
		cookie := sessionCookieOf(t, rec)
		if cookie.Value != "session-token" || !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/api" {
			t.Errorf("cookie = %+v, want HttpOnly, Secure, SameSite=Lax, Path=/api (ADR-0014)", cookie)
		}
		var body struct {
			ExpiresAt time.Time `json:"expires_at"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil || !body.ExpiresAt.Equal(testNow.Add(identity.SessionLifetime)) {
			t.Errorf("expires_at = %v, %v", body.ExpiresAt, err)
		}
		if strings.Contains(rec.Body.String(), "session-token") {
			t.Error("the session token is in the body; it belongs only in the cookie")
		}
	})

	t.Run("rejects a blank LINE user ID", func(t *testing.T) {
		rec := serve(newAuthRouter(true), http.MethodPost, "/api/v1/dev/login", `{"line_user_id":" ","display_name":"Taro"}`, nil)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"invalid_request"`) {
			t.Errorf("status = %d, body %s, want 400 invalid_request", rec.Code, rec.Body)
		}
	})
}

func TestMeAndLogout(t *testing.T) {
	router := newAuthRouter(true)

	t.Run("me without a session is 401 unauthenticated", func(t *testing.T) {
		rec := serve(router, http.MethodGet, "/api/v1/me", "", nil)
		if rec.Code != http.StatusUnauthorized || rec.Header().Get("Content-Type") != "application/problem+json" || !strings.Contains(rec.Body.String(), `"unauthenticated"`) {
			t.Errorf("status = %d, %s, body %s", rec.Code, rec.Header().Get("Content-Type"), rec.Body)
		}
	})

	t.Run("me with an unknown session is 401 unauthenticated", func(t *testing.T) {
		rec := serve(router, http.MethodGet, "/api/v1/me", "", &http.Cookie{Name: middleware.SessionCookieName, Value: "unknown", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})

	login := serve(router, http.MethodPost, "/api/v1/dev/login", `{"line_user_id":"U-organizer","display_name":"Taro"}`, nil)
	cookie := sessionCookieOf(t, login)

	t.Run("me returns the caller", func(t *testing.T) {
		rec := serve(router, http.MethodGet, "/api/v1/me", "", cookie)
		want := `{"display_name":"Taro","is_organizer":true,"roster_member":null,"status":"none"}`
		if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != want {
			t.Errorf("status = %d, body %s, want %s", rec.Code, rec.Body, want)
		}
	})

	t.Run("logout clears the cookie and ends the session", func(t *testing.T) {
		rec := serve(router, http.MethodDelete, "/api/v1/session", "", cookie)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
		}
		if cleared := sessionCookieOf(t, rec); cleared.MaxAge >= 0 || cleared.Value != "" {
			t.Errorf("cookie = %+v, want it cleared", cleared)
		}
		if rec := serve(router, http.MethodGet, "/api/v1/me", "", cookie); rec.Code != http.StatusUnauthorized {
			t.Errorf("me after logout: status = %d, want 401", rec.Code)
		}
	})
}
