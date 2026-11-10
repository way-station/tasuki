//go:build integration

package postgres

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestUnitOfWorkAfterCommit(t *testing.T) {
	client := newTestClient(t)
	unitOfWork := NewUnitOfWork(client)
	ctx := context.Background()

	t.Run("runs the hooks after the outer commit, in order", func(t *testing.T) {
		var ran []string
		err := unitOfWork.Run(ctx, func(ctx context.Context) error {
			afterCommit(ctx, func(context.Context) { ran = append(ran, "outer") })
			return unitOfWork.Run(ctx, func(ctx context.Context) error {
				afterCommit(ctx, func(context.Context) { ran = append(ran, "inner") })
				if len(ran) != 0 {
					t.Error("a hook ran before the commit")
				}
				return nil
			})
		})
		if err != nil || !slices.Equal(ran, []string{"outer", "inner"}) {
			t.Errorf("Run error = %v, hooks ran %v, want outer then inner", err, ran)
		}
	})

	t.Run("does not run the hooks when the work fails", func(t *testing.T) {
		ran := false
		failure := errors.New("roll back")
		err := unitOfWork.Run(ctx, func(ctx context.Context) error {
			afterCommit(ctx, func(context.Context) { ran = true })
			return failure
		})
		if !errors.Is(err, failure) || ran {
			t.Errorf("Run error = %v, hook ran = %v, want the failure and no hook", err, ran)
		}
	})

	t.Run("reports false outside a unit of work", func(t *testing.T) {
		if afterCommit(ctx, func(context.Context) {}) {
			t.Error("afterCommit outside a unit of work = true")
		}
	})
}
