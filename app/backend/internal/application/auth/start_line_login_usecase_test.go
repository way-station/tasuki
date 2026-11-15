package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fakeInviteLinkRepository struct {
	byHash map[string]*identity.InviteLink
	saved  []*identity.InviteLink
}

func (fake *fakeInviteLinkRepository) FindByTokenHash(_ context.Context, hash []byte) (*identity.InviteLink, error) {
	if link, ok := fake.byHash[string(hash)]; ok {
		return link, nil
	}
	return nil, shared.ErrNotFound
}

func (fake *fakeInviteLinkRepository) FindOpenByMember(context.Context, shared.MemberID) (*identity.InviteLink, error) {
	return nil, shared.ErrNotFound
}

func (fake *fakeInviteLinkRepository) Save(_ context.Context, link *identity.InviteLink) error {
	fake.saved = append(fake.saved, link)
	return nil
}

type fakeLoginAttemptRepository struct {
	byHash map[string]*identity.LoginAttempt
	saved  []*identity.LoginAttempt
}

func (fake *fakeLoginAttemptRepository) Save(_ context.Context, attempt *identity.LoginAttempt) error {
	fake.saved = append(fake.saved, attempt)
	return nil
}

func (fake *fakeLoginAttemptRepository) Take(_ context.Context, stateHash []byte) (*identity.LoginAttempt, error) {
	attempt, ok := fake.byHash[string(stateHash)]
	if !ok {
		return nil, shared.ErrNotFound
	}
	delete(fake.byHash, string(stateHash))
	return attempt, nil
}

type fakeLineLogin struct {
	identity                       LineIdentity
	err                            error
	gotCode, gotVerifier, gotNonce string
}

func (fake *fakeLineLogin) AuthorizationURL(state, nonce, codeVerifier string) string {
	return "https://access.line.me/authorize?state=" + state + "&nonce=" + nonce + "&verifier=" + codeVerifier
}

func (fake *fakeLineLogin) Authenticate(_ context.Context, code, codeVerifier, nonce string) (LineIdentity, error) {
	fake.gotCode, fake.gotVerifier, fake.gotNonce = code, codeVerifier, nonce
	return fake.identity, fake.err
}

func newStartLineLoginUseCase(inviteLinks *fakeInviteLinkRepository, attempts *fakeLoginAttemptRepository) *StartLineLoginUseCase {
	return NewStartLineLoginUseCase(inviteLinks, attempts, identity.NewLoginAttemptFactory(&sequentialIDs{}, fixedToken("random")), &fakeLineLogin{}, clock)
}

func TestStartLineLoginUseCase(t *testing.T) {
	t.Run("without an invite link", func(t *testing.T) {
		attempts := &fakeLoginAttemptRepository{}
		got, err := newStartLineLoginUseCase(&fakeInviteLinkRepository{}, attempts).Execute(context.Background(), StartLineLoginCommandFactory{}.Create(nil))
		if err != nil {
			t.Fatal(err)
		}
		if got != "https://access.line.me/authorize?state=random&nonce=random&verifier=random" {
			t.Errorf("Execute() = %q, want the URL of the state, the nonce, and the verifier", got)
		}
		if len(attempts.saved) != 1 {
			t.Fatalf("saved %d attempts, want 1", len(attempts.saved))
		}
		attempt := attempts.saved[0]
		if _, ok := attempt.InviteLinkID(); ok {
			t.Error("the attempt has an invite link, want none")
		}
		if string(attempt.StateHash()) != string(identity.HashToken("random")) || !attempt.ExpiresAt().Equal(now.Add(10*time.Minute)) {
			t.Error("the attempt does not keep the state's hash for 10 minutes")
		}
	})

	t.Run("from a usable invite link", func(t *testing.T) {
		invite := newInviteLink(t)
		attempts := &fakeLoginAttemptRepository{}
		inviteLinks := &fakeInviteLinkRepository{byHash: map[string]*identity.InviteLink{string(invite.TokenHash()): invite}}
		token := "invite-token"
		if _, err := newStartLineLoginUseCase(inviteLinks, attempts).Execute(context.Background(), StartLineLoginCommandFactory{}.Create(&token)); err != nil {
			t.Fatal(err)
		}
		if len(attempts.saved) != 1 {
			t.Fatalf("saved %d attempts, want 1", len(attempts.saved))
		}
		if got, ok := attempts.saved[0].InviteLinkID(); !ok || got != invite.ID() {
			t.Errorf("InviteLinkID() = %q, %v, want %q", got, ok, invite.ID())
		}
	})

	replaced := newInviteLink(t)
	replaced.Revoke(now)
	tests := []struct {
		name   string
		invite *identity.InviteLink
		token  string
	}{
		{name: "unknown invite token", token: "unknown"},
		{name: "blank invite token", token: ""},
		{name: "replaced invite link", invite: replaced, token: "invite-token"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			inviteLinks := &fakeInviteLinkRepository{byHash: map[string]*identity.InviteLink{}}
			if testCase.invite != nil {
				inviteLinks.byHash[string(testCase.invite.TokenHash())] = testCase.invite
			}
			attempts := &fakeLoginAttemptRepository{}
			_, err := newStartLineLoginUseCase(inviteLinks, attempts).Execute(context.Background(), StartLineLoginCommandFactory{}.Create(&testCase.token))
			if !errors.Is(err, identity.ErrInviteLinkUnusable) {
				t.Errorf("Execute() error = %v, want ErrInviteLinkUnusable", err)
			}
			if len(attempts.saved) != 0 {
				t.Error("an attempt was saved")
			}
		})
	}
}
