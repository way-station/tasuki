// Package auth holds the use cases of logging in and identifying the caller
// (ADR-0007, ADR-0009, ADR-0014).
package auth

import "errors"

// ErrUnauthenticated means the request has no usable session: no token, or one that is
// unknown, expired, or revoked. The API returns it as 401 unauthenticated.
var ErrUnauthenticated = errors.New("unauthenticated")

// ErrInvalidLoginCode means the one-time code is unknown, expired, or already exchanged.
// The API returns it as 400 invalid_request.
var ErrInvalidLoginCode = errors.New("invalid login code")

// ErrLoginFailed means LINE Login failed or was cancelled, or LINE called back for a login
// that is unknown, already finished, or older than identity.LoginAttemptLifetime. The
// callback redirects with error=login_failed.
var ErrLoginFailed = errors.New("login failed")

// ErrMemberLeft means the roster member of the user, or of the invite link the login
// started from, has left. The callback redirects with error=left and creates no session.
var ErrMemberLeft = errors.New("roster member left")
