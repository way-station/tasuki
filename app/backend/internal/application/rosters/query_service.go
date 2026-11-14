package rosters

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RosterQueryService reads the roster with what the roster screens show. It is implemented
// in the infrastructure layer.
type RosterQueryService interface {
	// ListRosterMembers returns every roster member of the team, by name, with their links
	// and their invite link that is neither used nor replaced.
	ListRosterMembers(ctx context.Context, teamID shared.TeamID) ([]RosterMemberRecord, error)
	// GetRosterMember returns the member of the team with their average time and their laps
	// by race. It returns an error wrapping shared.ErrNotFound when the team has no such
	// member.
	GetRosterMember(ctx context.Context, teamID shared.TeamID, memberID shared.MemberID) (RosterMemberDetailRecord, error)
	// ListPendingLinks returns the links of the team's members that wait for approval,
	// oldest first.
	ListPendingLinks(ctx context.Context, teamID shared.TeamID) ([]PendingLinkRecord, error)
	// ListInvitedRaces returns the races the member is invited to, with their dates.
	ListInvitedRaces(ctx context.Context, memberID shared.MemberID) ([]roster.InvitedRace, error)
}

// RosterMemberRecord is a roster member as ListRosterMembers reads them.
type RosterMemberRecord struct {
	MemberID   shared.MemberID
	Name       string
	Left       bool
	Links      []LinkRecord
	InviteLink *InviteLinkRecord // nil when there is no open invite link
}

// LinkRecord is a link between a user and the roster member.
type LinkRecord struct {
	ID           shared.MemberLinkID
	UserID       shared.UserID
	InviteLinkID shared.InviteLinkID
	CreatedAt    time.Time
	ApprovedAt   *time.Time
	RejectedAt   *time.Time
	UnlinkedAt   *time.Time
}

// InviteLinkRecord is the invite link of a roster member that is neither used nor replaced.
type InviteLinkRecord struct {
	ID        shared.InviteLinkID
	ExpiresAt time.Time
}

// RosterMemberDetailRecord is a roster member as GetRosterMember reads them.
type RosterMemberDetailRecord struct {
	MemberID   shared.MemberID
	Name       string
	Left       bool
	Average    shared.Pace
	LapCount   int
	LapsByRace []RaceLapsRecord
}

// RaceLapsRecord is the member's laps for one race, in submission order.
type RaceLapsRecord struct {
	RaceID   shared.RaceID
	RaceName string
	Laps     []LapRecord
}

// LapRecord is one lap.
type LapRecord struct {
	ID   shared.LapID
	Time shared.Pace
}

// PendingLinkRecord is a link that waits for approval, as ListPendingLinks reads it.
type PendingLinkRecord struct {
	LinkID          shared.MemberLinkID
	LineDisplayName string
	MemberID        shared.MemberID
	MemberName      string
	CreatedAt       time.Time
}
