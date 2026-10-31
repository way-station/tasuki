package random

import (
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
)

func TestIDGenerator(t *testing.T) {
	id, err := uuid.Parse(IDGenerator{}.NewID())
	if err != nil || id.Version() != 7 {
		t.Errorf("NewID() = %v, %v, want a UUIDv7", id, err)
	}
}

func TestTokenGenerator(t *testing.T) {
	first, err := TokenGenerator{}.NewToken()
	if err != nil {
		t.Fatal(err)
	}
	second, _ := TokenGenerator{}.NewToken()

	decoded, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil || len(decoded) != 32 {
		t.Errorf("NewToken() = %q, want 32 random bytes in URL-safe base64", first)
	}
	if first == second {
		t.Error("two tokens are equal")
	}
}
