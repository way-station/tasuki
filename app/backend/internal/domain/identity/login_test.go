package identity

import (
	"bytes"
	"errors"
	"strconv"
	"testing"
	"time"
)

// seqTokens returns token1, token2, ... so that tokens of one value can be told apart.
type seqTokens struct{ count int }

func (tokens *seqTokens) NewToken() (string, error) {
	tokens.count++
	return "token" + strconv.Itoa(tokens.count), nil
}

func TestLoginAttempt(t *testing.T) {
	t.Run("keeps the state's hash, a nonce, and a PKCE verifier, for 10 minutes", func(t *testing.T) {
		attempt, state, err := NewLoginAttemptFactory(&seqIDs{}, &seqTokens{}).New(nil, now)
		if err != nil {
			t.Fatal(err)
		}

		if state != "token1" || !bytes.Equal(attempt.StateHash(), HashToken("token1")) {
			t.Errorf("state = %q, hash does not match", state)
		}
		if attempt.Nonce() != "token2" || attempt.CodeVerifier() != "token3" {
			t.Errorf("Nonce() = %q, CodeVerifier() = %q, want two other tokens", attempt.Nonce(), attempt.CodeVerifier())
		}
		if _, ok := attempt.InviteLinkID(); ok {
			t.Error("InviteLinkID() is set without an invite link")
		}
		if want := now.Add(10 * time.Minute); !attempt.ExpiresAt().Equal(want) {
			t.Errorf("ExpiresAt() = %v, want %v", attempt.ExpiresAt(), want)
		}
		if !attempt.Usable(attempt.ExpiresAt().Add(-time.Second)) || attempt.Usable(attempt.ExpiresAt()) {
			t.Error("Usable() must hold until the attempt expires, and not after")
		}
	})

	t.Run("remembers a usable invite link", func(t *testing.T) {
		inviteLink, _ := newInvite(t)

		attempt, _, err := NewLoginAttemptFactory(&seqIDs{}, &seqTokens{}).New(inviteLink, now)
		if err != nil {
			t.Fatal(err)
		}

		if id, ok := attempt.InviteLinkID(); !ok || id != inviteLink.ID() {
			t.Errorf("InviteLinkID() = %q, %v, want %q", id, ok, inviteLink.ID())
		}
	})

	t.Run("refuses an invite link that cannot be used", func(t *testing.T) {
		inviteLink, _ := newInvite(t)

		_, _, err := NewLoginAttemptFactory(&seqIDs{}, &seqTokens{}).New(inviteLink, inviteLink.ExpiresAt())

		if !errors.Is(err, ErrInviteLinkUnusable) {
			t.Errorf("New() error = %v, want ErrInviteLinkUnusable", err)
		}
	})
}

func TestLoginCode(t *testing.T) {
	code, plain, err := NewLoginCodeFactory(&seqIDs{}, fixedToken("one-time")).New("a1", now)
	if err != nil {
		t.Fatal(err)
	}

	if plain != "one-time" || !bytes.Equal(code.CodeHash(), HashToken("one-time")) {
		t.Errorf("code = %q, hash does not match", plain)
	}
	if code.UserID() != "a1" {
		t.Errorf("UserID() = %q, want a1", code.UserID())
	}
	if want := now.Add(time.Minute); !code.ExpiresAt().Equal(want) {
		t.Errorf("ExpiresAt() = %v, want %v", code.ExpiresAt(), want)
	}
	if !code.Usable(code.ExpiresAt().Add(-time.Second)) || code.Usable(code.ExpiresAt()) {
		t.Error("Usable() must hold until the code expires, and not after")
	}
}
