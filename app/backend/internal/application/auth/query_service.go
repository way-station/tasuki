package auth

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// SessionQueryService reads sessions to identify the caller.
// It is implemented in the infrastructure layer.
type SessionQueryService interface {
	// FindActiveAccountID returns the account of the session whose token hash is hash and
	// that is neither expired nor revoked at now. It returns an error wrapping
	// shared.ErrNotFound when there is none.
	FindActiveAccountID(ctx context.Context, hash []byte, now time.Time) (shared.AccountID, error)
}

// MeQueryService reads what GET /me needs. It is implemented in the infrastructure layer.
type MeQueryService interface {
	// FindMe returns the account and all of its member links. It returns an error wrapping
	// shared.ErrNotFound when the account does not exist.
	FindMe(ctx context.Context, accountID shared.AccountID) (MeRecord, error)
}

// MeRecord is an account as MeQueryService reads it.
type MeRecord struct {
	LineUserID  string
	DisplayName string
	Links       []MeLinkRecord
}

// MeLinkRecord is a member link of the account, with its roster member's name.
type MeLinkRecord struct {
	ID               shared.MemberLinkID
	RosterMemberID   shared.MemberID
	RosterMemberName string
	InviteLinkID     shared.InviteLinkID
	CreatedAt        time.Time
	ApprovedAt       *time.Time
	RejectedAt       *time.Time
	UnlinkedAt       *time.Time
}
