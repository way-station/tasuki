package auth

import "context"

// LineLogin runs LINE Login (OpenID Connect with PKCE, ADR-0007) on behalf of the use
// cases. It is implemented in the infrastructure layer, so how LINE is called stays out of
// the use cases.
type LineLogin interface {
	// AuthorizationURL returns the URL of LINE's login page. state comes back on the
	// callback, nonce comes back in the ID token, and codeVerifier is the PKCE verifier
	// that Authenticate must be given later.
	AuthorizationURL(state, nonce, codeVerifier string) string
	// Authenticate exchanges the authorization code from the callback for an ID token,
	// verifies it, and returns the LINE user it names. It fails when the exchange fails
	// or the ID token is not valid for nonce.
	Authenticate(ctx context.Context, code, codeVerifier, nonce string) (LineIdentity, error)
}

// LineIdentity is a LINE user whose ID token was verified.
type LineIdentity struct {
	// UserID is the sub claim, the LINE user ID.
	UserID      string
	DisplayName string
}
