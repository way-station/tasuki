package identity

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LoginAttemptLifetime is how long a LINE Login may take, from starting it to LINE's
// callback (docs/design/database.md).
const LoginAttemptLifetime = 10 * time.Minute

// LoginAttempt is a LINE Login that a browser started and LINE has not called back for
// yet (ADR-0007). The browser takes the state to LINE and brings it back; the attempt
// keeps only its hash, with the nonce and the PKCE verifier that finish the login.
// It is created by LoginAttemptFactory.
type LoginAttempt struct {
	id           shared.LoginAttemptID
	stateHash    []byte
	nonce        string
	codeVerifier string
	inviteLinkID *shared.InviteLinkID
	expiresAt    time.Time
}

// ID returns the attempt's ID.
func (attempt *LoginAttempt) ID() shared.LoginAttemptID { return attempt.id }

// StateHash returns the SHA-256 hash of the state.
func (attempt *LoginAttempt) StateHash() []byte { return attempt.stateHash }

// Nonce returns the nonce the ID token must carry.
func (attempt *LoginAttempt) Nonce() string { return attempt.nonce }

// CodeVerifier returns the PKCE verifier sent with the authorization code.
func (attempt *LoginAttempt) CodeVerifier() string { return attempt.codeVerifier }

// InviteLinkID returns the invite link the login started from, and false if it started
// without one.
func (attempt *LoginAttempt) InviteLinkID() (shared.InviteLinkID, bool) {
	if attempt.inviteLinkID == nil {
		return "", false
	}
	return *attempt.inviteLinkID, true
}

// ExpiresAt returns when the attempt can no longer be finished.
func (attempt *LoginAttempt) ExpiresAt() time.Time { return attempt.expiresAt }

// Usable reports whether the login can still be finished at now.
func (attempt *LoginAttempt) Usable(now time.Time) bool {
	return now.Before(attempt.expiresAt)
}

// LoginAttemptFactory assembles LoginAttempt values.
type LoginAttemptFactory struct {
	ids    shared.IDGenerator
	tokens TokenGenerator
}

// NewLoginAttemptFactory returns a LoginAttemptFactory.
func NewLoginAttemptFactory(ids shared.IDGenerator, tokens TokenGenerator) LoginAttemptFactory {
	return LoginAttemptFactory{ids: ids, tokens: tokens}
}

// New returns a new attempt starting at now, and its state. inviteLink is the link the
// login starts from, or nil; it fails with ErrInviteLinkUnusable when the link cannot be
// used. The state is returned only here; it goes to LINE and is never stored.
func (factory LoginAttemptFactory) New(inviteLink *InviteLink, now time.Time) (*LoginAttempt, string, error) {
	var inviteLinkID *shared.InviteLinkID
	if inviteLink != nil {
		if !inviteLink.Usable(now) {
			return nil, "", ErrInviteLinkUnusable
		}
		id := inviteLink.ID()
		inviteLinkID = &id
	}
	state, err := factory.tokens.NewToken()
	if err != nil {
		return nil, "", err
	}
	nonce, err := factory.tokens.NewToken()
	if err != nil {
		return nil, "", err
	}
	codeVerifier, err := factory.tokens.NewToken()
	if err != nil {
		return nil, "", err
	}
	return &LoginAttempt{
		id:           shared.LoginAttemptID(factory.ids.NewID()),
		stateHash:    HashToken(state),
		nonce:        nonce,
		codeVerifier: codeVerifier,
		inviteLinkID: inviteLinkID,
		expiresAt:    now.Add(LoginAttemptLifetime),
	}, state, nil
}

// Reconstruct returns an attempt read from storage.
func (factory LoginAttemptFactory) Reconstruct(id shared.LoginAttemptID, stateHash []byte, nonce, codeVerifier string, inviteLinkID *shared.InviteLinkID, expiresAt time.Time) *LoginAttempt {
	return &LoginAttempt{id: id, stateHash: stateHash, nonce: nonce, codeVerifier: codeVerifier, inviteLinkID: inviteLinkID, expiresAt: expiresAt}
}
