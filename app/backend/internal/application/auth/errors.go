// Package auth holds the use cases of logging in and identifying the caller
// (ADR-0007, ADR-0009, ADR-0014).
package auth

import "errors"

// ErrUnauthenticated means the request has no usable session: no token, or one that is
// unknown, expired, or revoked. The API returns it as 401 unauthenticated.
var ErrUnauthenticated = errors.New("unauthenticated")
