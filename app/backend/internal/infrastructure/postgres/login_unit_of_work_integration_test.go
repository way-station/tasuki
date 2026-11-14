//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

// TestLoginRepositoriesInUnitOfWork checks that the user and the one-time code that LINE
// Login's callback saves are rolled back with the unit of work.
func TestLoginRepositoriesInUnitOfWork(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	userFactory := identity.NewUserFactory(random.IDGenerator{})
	users := NewUserRepository(client, userFactory)
	codeFactory := identity.NewLoginCodeFactory(random.IDGenerator{}, random.TokenGenerator{})
	codes := NewLoginCodeRepository(client, codeFactory)
	errStop := errors.New("stop")

	var plain string
	err := NewUnitOfWork(client).Run(ctx, func(ctx context.Context) error {
		user, err := userFactory.New("U-rolled-back", "Taro")
		if err != nil {
			return err
		}
		if err := users.Save(ctx, user); err != nil {
			return err
		}
		if _, err := users.FindByLineUserID(ctx, "U-rolled-back"); err != nil {
			t.Errorf("FindByLineUserID in the unit of work: %v", err)
		}
		code, codePlain, err := codeFactory.New(user.ID(), testNow)
		if err != nil {
			return err
		}
		plain = codePlain
		if err := codes.Save(ctx, code); err != nil {
			return err
		}
		return errStop
	})
	if !errors.Is(err, errStop) {
		t.Fatalf("Run() error = %v, want the error of the work", err)
	}
	if _, err := users.FindByLineUserID(ctx, "U-rolled-back"); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByLineUserID after the rollback error = %v, want ErrNotFound", err)
	}
	if _, err := codes.Take(ctx, identity.HashToken(plain)); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Take after the rollback error = %v, want ErrNotFound", err)
	}
}
