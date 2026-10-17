package roster

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MemberRepository stores roster members. It is implemented in the infrastructure layer.
type MemberRepository interface {
	// FindByID returns the member, or an error wrapping shared.ErrNotFound.
	FindByID(ctx context.Context, id shared.MemberID) (*Member, error)
	// Save inserts or updates the member.
	Save(ctx context.Context, m *Member) error
}
