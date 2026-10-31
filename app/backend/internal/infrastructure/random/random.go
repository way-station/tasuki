// Package random creates IDs and secret tokens for the domain's factories.
package random

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

// IDGenerator implements shared.IDGenerator with UUIDv7, so new rows sort by time
// (docs/design/database.md).
type IDGenerator struct{}

// NewID returns a new UUIDv7.
func (IDGenerator) NewID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// tokenBytes is the size of a token: 256 bits from a cryptographically secure generator.
const tokenBytes = 32

// TokenGenerator implements identity.TokenGenerator for invite links and sessions.
type TokenGenerator struct{}

// NewToken returns a random token, URL-safe so it fits in a URL and a cookie.
func (TokenGenerator) NewToken() (string, error) {
	buffer := make([]byte, tokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
