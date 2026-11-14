package auth

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// InviteLinkQueryService reads invite links with the name of the team of their roster
// member. It is implemented in the infrastructure layer.
type InviteLinkQueryService interface {
	// FindByTokenHash returns the invite link whose token has hash. It returns an error
	// wrapping shared.ErrNotFound when there is none.
	FindByTokenHash(ctx context.Context, hash []byte) (InviteLinkRecord, error)
	// FindByID returns the invite link. It returns an error wrapping shared.ErrNotFound
	// when there is none.
	FindByID(ctx context.Context, id shared.InviteLinkID) (InviteLinkRecord, error)
}

// InviteLinkRecord is an invite link as InviteLinkQueryService reads it.
type InviteLinkRecord struct {
	ID        shared.InviteLinkID
	MemberID  shared.MemberID
	TokenHash []byte
	ExpiresAt time.Time
	UsedAt    *time.Time
	RevokedAt *time.Time
	TeamName  string
}
