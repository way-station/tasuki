package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fakeLoginCodeRepository struct {
	byHash map[string]*identity.LoginCode
	saved  []*identity.LoginCode
}

func (fake *fakeLoginCodeRepository) Save(_ context.Context, code *identity.LoginCode) error {
	fake.saved = append(fake.saved, code)
	return nil
}

func (fake *fakeLoginCodeRepository) Take(_ context.Context, codeHash []byte) (*identity.LoginCode, error) {
	code, ok := fake.byHash[string(codeHash)]
	if !ok {
		return nil, shared.ErrNotFound
	}
	delete(fake.byHash, string(codeHash))
	return code, nil
}

func newLoginCodes(t *testing.T) *fakeLoginCodeRepository {
	t.Helper()
	code, _, err := identity.NewLoginCodeFactory(&sequentialIDs{}, fixedToken("login-code")).New("u1", now)
	if err != nil {
		t.Fatal(err)
	}
	return &fakeLoginCodeRepository{byHash: map[string]*identity.LoginCode{string(code.CodeHash()): code}}
}

func newExchangeLoginCodeUseCase(codes *fakeLoginCodeRepository, sessions *fakeSessionRepository, at func() time.Time) *ExchangeLoginCodeUseCase {
	return NewExchangeLoginCodeUseCase(codes, sessions, identity.NewSessionFactory(&sequentialIDs{}, fixedToken("session-token")), SessionDTOFactory{}, at)
}

func TestExchangeLoginCodeUseCase(t *testing.T) {
	t.Run("creates a session for the user of the code", func(t *testing.T) {
		sessions := &fakeSessionRepository{}
		got, err := newExchangeLoginCodeUseCase(newLoginCodes(t), sessions, clock).Execute(context.Background(), "login-code")
		if err != nil {
			t.Fatal(err)
		}
		if got.Token() != "session-token" {
			t.Errorf("Token() = %q, want session-token", got.Token())
		}
		if len(sessions.saved) != 1 || sessions.saved[0].UserID() != "u1" || !got.ExpiresAt().Equal(sessions.saved[0].ExpiresAt()) {
			t.Error("the session of u1 was not saved with the expiry returned")
		}
	})

	t.Run("a code can be exchanged once", func(t *testing.T) {
		useCase := newExchangeLoginCodeUseCase(newLoginCodes(t), &fakeSessionRepository{}, clock)
		if _, err := useCase.Execute(context.Background(), "login-code"); err != nil {
			t.Fatal(err)
		}
		if _, err := useCase.Execute(context.Background(), "login-code"); !errors.Is(err, ErrInvalidLoginCode) {
			t.Errorf("second Execute() error = %v, want ErrInvalidLoginCode", err)
		}
	})

	tests := []struct {
		name string
		code string
		at   func() time.Time
	}{
		{name: "unknown code", code: "unknown", at: clock},
		{name: "expired code", code: "login-code", at: func() time.Time { return now.Add(identity.LoginCodeLifetime) }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			sessions := &fakeSessionRepository{}
			_, err := newExchangeLoginCodeUseCase(newLoginCodes(t), sessions, testCase.at).Execute(context.Background(), testCase.code)
			if !errors.Is(err, ErrInvalidLoginCode) {
				t.Errorf("Execute() error = %v, want ErrInvalidLoginCode", err)
			}
			if len(sessions.saved) != 0 {
				t.Error("a session was saved")
			}
		})
	}
}
