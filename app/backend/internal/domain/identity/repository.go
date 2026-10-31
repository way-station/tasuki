package identity

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// UserRepository stores users. Errors for missing users wrap shared.ErrNotFound.
type UserRepository interface {
	FindByID(ctx context.Context, id shared.UserID) (*User, error)
	FindByLineUserID(ctx context.Context, lineUserID string) (*User, error)
	Save(ctx context.Context, a *User) error
}

// InviteLinkRepository stores invite links. Errors for missing links wrap shared.ErrNotFound.
type InviteLinkRepository interface {
	FindByTokenHash(ctx context.Context, hash []byte) (*InviteLink, error)
	// FindOpenByMember returns the member's link that is neither used nor replaced.
	FindOpenByMember(ctx context.Context, memberID shared.MemberID) (*InviteLink, error)
	Save(ctx context.Context, l *InviteLink) error
}

// MemberLinkRepository stores links between users and roster members.
// Errors for missing links wrap shared.ErrNotFound.
type MemberLinkRepository interface {
	FindByID(ctx context.Context, id shared.MemberLinkID) (*MemberLink, error)
	ListByUser(ctx context.Context, userID shared.UserID) ([]*MemberLink, error)
	ListByMember(ctx context.Context, memberID shared.MemberID) ([]*MemberLink, error)
	Save(ctx context.Context, l *MemberLink) error
}

// SessionRepository stores login sessions. Errors for missing sessions wrap shared.ErrNotFound.
type SessionRepository interface {
	FindByTokenHash(ctx context.Context, hash []byte) (*Session, error)
	Save(ctx context.Context, session *Session) error
}
