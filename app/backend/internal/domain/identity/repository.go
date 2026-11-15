package identity

import (
	"context"
	"time"

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

// UserSessionRepository lists the sessions of a user, so that they can all be revoked when
// the user's roster member leaves (ADR-0014). The revoked sessions are saved with
// SessionRepository.
type UserSessionRepository interface {
	// ListActiveByUser returns the user's sessions that are active at now.
	ListActiveByUser(ctx context.Context, userID shared.UserID, now time.Time) ([]*Session, error)
}

// LoginAttemptRepository stores LINE Logins in progress. Errors for missing attempts wrap
// shared.ErrNotFound.
type LoginAttemptRepository interface {
	Save(ctx context.Context, attempt *LoginAttempt) error
	// Take deletes the attempt whose state has hash and returns it, expired or not, so the
	// same state can be taken only once.
	Take(ctx context.Context, stateHash []byte) (*LoginAttempt, error)
}

// LoginCodeRepository stores one-time codes. Errors for missing codes wrap
// shared.ErrNotFound.
type LoginCodeRepository interface {
	Save(ctx context.Context, code *LoginCode) error
	// Take deletes the code whose hash is hash and returns it, expired or not, so the same
	// code can be taken only once.
	Take(ctx context.Context, codeHash []byte) (*LoginCode, error)
}
