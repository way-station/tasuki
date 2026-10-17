package identity

import (
	"crypto/sha256"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// InviteLinkLifetime is how long an invite link can be used (ADR-0006).
const InviteLinkLifetime = 7 * 24 * time.Hour

// ErrInviteLinkUnusable means the invite link is expired, used, or replaced.
// The API returns it as invite_unusable.
var ErrInviteLinkUnusable = errors.New("invite link unusable")

// InviteLink is a personal link an organizer creates for one roster member. It can be used
// once, for one week, and a newer link for the same member replaces it.
// It is created by InviteLinkFactory. Only the hash of the token is kept.
type InviteLink struct {
	id        shared.InviteLinkID
	memberID  shared.MemberID
	tokenHash []byte
	expiresAt time.Time
	usedAt    *time.Time
	revokedAt *time.Time
}

// ID returns the link's ID.
func (l *InviteLink) ID() shared.InviteLinkID { return l.id }

// MemberID returns the roster member the link is for.
func (l *InviteLink) MemberID() shared.MemberID { return l.memberID }

// TokenHash returns the SHA-256 hash of the token.
func (l *InviteLink) TokenHash() []byte { return l.tokenHash }

// ExpiresAt returns when the link expires.
func (l *InviteLink) ExpiresAt() time.Time { return l.expiresAt }

// UsedAt returns when the link was used, and false if it has not been used.
func (l *InviteLink) UsedAt() (time.Time, bool) { return deref(l.usedAt) }

// RevokedAt returns when a newer link replaced this one, and false if it has not been replaced.
func (l *InviteLink) RevokedAt() (time.Time, bool) { return deref(l.revokedAt) }

// Usable reports whether the link can still be used at now.
func (l *InviteLink) Usable(now time.Time) bool {
	return l.usedAt == nil && l.revokedAt == nil && now.Before(l.expiresAt)
}

// use marks the link as used. Only MemberLinkFactory calls it, when the link creates a
// member link.
func (l *InviteLink) use(now time.Time) error {
	if !l.Usable(now) {
		return ErrInviteLinkUnusable
	}
	l.usedAt = &now
	return nil
}

// Revoke makes the link unusable because a newer link replaces it.
func (l *InviteLink) Revoke(now time.Time) {
	if l.usedAt == nil && l.revokedAt == nil {
		l.revokedAt = &now
	}
}

// TokenGenerator creates the random token of an invite link. It is implemented in the
// infrastructure layer.
type TokenGenerator interface {
	NewToken() (string, error)
}

// HashToken returns the hash under which a token is stored and looked up.
func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// InviteLinkFactory assembles InviteLink values.
type InviteLinkFactory struct {
	ids    shared.IDGenerator
	tokens TokenGenerator
}

// NewInviteLinkFactory returns an InviteLinkFactory.
func NewInviteLinkFactory(ids shared.IDGenerator, tokens TokenGenerator) InviteLinkFactory {
	return InviteLinkFactory{ids: ids, tokens: tokens}
}

// New returns a new invite link for the member, valid from now, and its token. The token is
// returned only here; it goes into the URL and is never stored.
// Check roster.Member.CanIssueInviteLink first, and revoke the member's previous link.
func (f InviteLinkFactory) New(memberID shared.MemberID, now time.Time) (*InviteLink, string, error) {
	token, err := f.tokens.NewToken()
	if err != nil {
		return nil, "", err
	}
	return &InviteLink{
		id:        shared.InviteLinkID(f.ids.NewID()),
		memberID:  memberID,
		tokenHash: HashToken(token),
		expiresAt: now.Add(InviteLinkLifetime),
	}, token, nil
}

// Reconstruct returns an invite link read from storage.
func (f InviteLinkFactory) Reconstruct(id shared.InviteLinkID, memberID shared.MemberID, tokenHash []byte, expiresAt time.Time, usedAt, revokedAt *time.Time) *InviteLink {
	return &InviteLink{id: id, memberID: memberID, tokenHash: tokenHash, expiresAt: expiresAt, usedAt: usedAt, revokedAt: revokedAt}
}

func deref(t *time.Time) (time.Time, bool) {
	if t == nil {
		return time.Time{}, false
	}
	return *t, true
}
