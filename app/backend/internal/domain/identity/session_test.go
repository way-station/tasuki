package identity

import (
	"bytes"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	session, token, err := NewSessionFactory(&seqIDs{}, fixedToken("session-secret")).New("a1", now)
	if err != nil {
		t.Fatal(err)
	}

	if token != "session-secret" || !bytes.Equal(session.TokenHash(), HashToken("session-secret")) {
		t.Errorf("token = %q, hash does not match", token)
	}
	if session.AccountID() != "a1" {
		t.Errorf("AccountID() = %q, want a1", session.AccountID())
	}
	if want := now.Add(30 * 24 * time.Hour); !session.ExpiresAt().Equal(want) {
		t.Errorf("ExpiresAt() = %v, want %v", session.ExpiresAt(), want)
	}

	if !session.Active(now) || !session.Active(session.ExpiresAt().Add(-time.Second)) {
		t.Error("Active() before it expires = false, want true")
	}
	if session.Active(session.ExpiresAt()) {
		t.Error("Active() when it expires = true, want false")
	}

	revokedAt := now.Add(time.Hour)
	session.Revoke(revokedAt)
	session.Revoke(revokedAt.Add(time.Hour))
	if at, ok := session.RevokedAt(); !ok || !at.Equal(revokedAt) {
		t.Errorf("RevokedAt() = %v, %v, want the first revocation %v", at, ok, revokedAt)
	}
	if session.Active(now) {
		t.Error("Active() after Revoke = true, want false")
	}
}
