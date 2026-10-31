package identity

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// SessionLifetime is how long a session lasts. There is no refresh token, so the user logs
// in with LINE again after it (ADR-0014).
const SessionLifetime = 30 * 24 * time.Hour

// Session is a login of an account, kept on the server (ADR-0014). The browser holds only
// the session token; the session keeps its SHA-256 hash, so stored sessions do not give
// usable tokens. It is created by SessionFactory.
type Session struct {
	id        shared.SessionID
	accountID shared.AccountID
	tokenHash []byte
	expiresAt time.Time
	revokedAt *time.Time
}

// ID returns the session's ID.
func (session *Session) ID() shared.SessionID { return session.id }

// AccountID returns the account that logged in.
func (session *Session) AccountID() shared.AccountID { return session.accountID }

// TokenHash returns the SHA-256 hash of the session token.
func (session *Session) TokenHash() []byte { return session.tokenHash }

// ExpiresAt returns when the session ends.
func (session *Session) ExpiresAt() time.Time { return session.expiresAt }

// RevokedAt returns when the session was revoked, and false if it was not.
func (session *Session) RevokedAt() (time.Time, bool) { return deref(session.revokedAt) }

// Active reports whether the session can still be used at now: not expired and not revoked.
func (session *Session) Active(now time.Time) bool {
	return session.revokedAt == nil && now.Before(session.expiresAt)
}

// Revoke ends the session, on logout or when the account's roster member leaves. Revoking
// a revoked session keeps the first time.
func (session *Session) Revoke(now time.Time) {
	if session.revokedAt == nil {
		session.revokedAt = &now
	}
}

// SessionFactory assembles Session values.
type SessionFactory struct {
	ids    shared.IDGenerator
	tokens TokenGenerator
}

// NewSessionFactory returns a SessionFactory.
func NewSessionFactory(ids shared.IDGenerator, tokens TokenGenerator) SessionFactory {
	return SessionFactory{ids: ids, tokens: tokens}
}

// New returns a new session for the account, starting at now, and its token. The token is
// returned only here; it goes into the cookie and is never stored.
func (factory SessionFactory) New(accountID shared.AccountID, now time.Time) (*Session, string, error) {
	token, err := factory.tokens.NewToken()
	if err != nil {
		return nil, "", err
	}
	return &Session{
		id:        shared.SessionID(factory.ids.NewID()),
		accountID: accountID,
		tokenHash: HashToken(token),
		expiresAt: now.Add(SessionLifetime),
	}, token, nil
}

// Reconstruct returns a session read from storage.
func (factory SessionFactory) Reconstruct(id shared.SessionID, accountID shared.AccountID, tokenHash []byte, expiresAt time.Time, revokedAt *time.Time) *Session {
	return &Session{id: id, accountID: accountID, tokenHash: tokenHash, expiresAt: expiresAt, revokedAt: revokedAt}
}
