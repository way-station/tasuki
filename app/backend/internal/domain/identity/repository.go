package identity

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// AccountRepository stores accounts. Errors for missing accounts wrap shared.ErrNotFound.
type AccountRepository interface {
	FindByID(ctx context.Context, id shared.AccountID) (*Account, error)
	FindByLineUserID(ctx context.Context, lineUserID string) (*Account, error)
	Save(ctx context.Context, a *Account) error
}

// InviteLinkRepository stores invite links. Errors for missing links wrap shared.ErrNotFound.
type InviteLinkRepository interface {
	FindByTokenHash(ctx context.Context, hash []byte) (*InviteLink, error)
	// FindOpenByMember returns the member's link that is neither used nor replaced.
	FindOpenByMember(ctx context.Context, memberID shared.MemberID) (*InviteLink, error)
	Save(ctx context.Context, l *InviteLink) error
}

// MemberLinkRepository stores links between accounts and roster members.
// Errors for missing links wrap shared.ErrNotFound.
type MemberLinkRepository interface {
	FindByID(ctx context.Context, id shared.MemberLinkID) (*MemberLink, error)
	ListByAccount(ctx context.Context, accountID shared.AccountID) ([]*MemberLink, error)
	ListByMember(ctx context.Context, memberID shared.MemberID) ([]*MemberLink, error)
	Save(ctx context.Context, l *MemberLink) error
}
