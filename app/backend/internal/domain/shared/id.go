// Package shared holds what every domain package uses: IDs, domain errors, dates, and times.
// It imports no other layer (ADR-0011).
package shared

// IDGenerator creates new IDs. Entity factories use it; it is implemented in the
// infrastructure layer (UUIDv7, docs/design/database.md).
type IDGenerator interface {
	NewID() string
}

// IDs of the entities are distinct types, so that one kind of ID cannot be passed where
// another is expected.

// TeamID identifies a team (tenant).
type TeamID string

// UserID identifies a user (a LINE login).
type UserID string

// MemberID identifies a roster member.
type MemberID string

// InviteLinkID identifies an invite link.
type InviteLinkID string

// MemberLinkID identifies a link between a user and a roster member.
type MemberLinkID string

// SessionID identifies a login session (ADR-0014).
type SessionID string

// RaceID identifies a race.
type RaceID string

// EntryID identifies an entry of a race.
type EntryID string

// RunningOrderID identifies a running order.
type RunningOrderID string

// LapID identifies a lap.
type LapID string
