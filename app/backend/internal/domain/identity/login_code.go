package identity

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LoginCodeLifetime is how long the frontend has to exchange a one-time code for a
// session (docs/design/database.md).
const LoginCodeLifetime = time.Minute

// LoginCode is a one-time code that LINE Login's callback hands to the frontend in the URL,
// so the session token itself never appears in a URL. POST /session exchanges it for a
// session. Only the hash of the code is kept. It is created by LoginCodeFactory.
type LoginCode struct {
	id        shared.LoginCodeID
	userID    shared.UserID
	codeHash  []byte
	expiresAt time.Time
}

// ID returns the code's ID.
func (code *LoginCode) ID() shared.LoginCodeID { return code.id }

// UserID returns the user that logged in.
func (code *LoginCode) UserID() shared.UserID { return code.userID }

// CodeHash returns the SHA-256 hash of the code.
func (code *LoginCode) CodeHash() []byte { return code.codeHash }

// ExpiresAt returns when the code can no longer be exchanged.
func (code *LoginCode) ExpiresAt() time.Time { return code.expiresAt }

// Usable reports whether the code can still be exchanged at now.
func (code *LoginCode) Usable(now time.Time) bool {
	return now.Before(code.expiresAt)
}

// LoginCodeFactory assembles LoginCode values.
type LoginCodeFactory struct {
	ids    shared.IDGenerator
	tokens TokenGenerator
}

// NewLoginCodeFactory returns a LoginCodeFactory.
func NewLoginCodeFactory(ids shared.IDGenerator, tokens TokenGenerator) LoginCodeFactory {
	return LoginCodeFactory{ids: ids, tokens: tokens}
}

// New returns a new code for the user, valid from now, and the code itself. The code is
// returned only here; it goes into the redirect URL and is never stored.
func (factory LoginCodeFactory) New(userID shared.UserID, now time.Time) (*LoginCode, string, error) {
	plain, err := factory.tokens.NewToken()
	if err != nil {
		return nil, "", err
	}
	return &LoginCode{
		id:        shared.LoginCodeID(factory.ids.NewID()),
		userID:    userID,
		codeHash:  HashToken(plain),
		expiresAt: now.Add(LoginCodeLifetime),
	}, plain, nil
}

// Reconstruct returns a code read from storage.
func (factory LoginCodeFactory) Reconstruct(id shared.LoginCodeID, userID shared.UserID, codeHash []byte, expiresAt time.Time) *LoginCode {
	return &LoginCode{id: id, userID: userID, codeHash: codeHash, expiresAt: expiresAt}
}
