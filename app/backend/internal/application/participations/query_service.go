package participations

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// InvitationQueryService reads the roster members together with their invitation to a
// race. It is implemented in the infrastructure layer.
type InvitationQueryService interface {
	// ListInvitations returns every roster member of the team who has not left, with
	// whether they are invited to the race and their answer. It returns an error wrapping
	// shared.ErrNotFound when the team has no such race.
	ListInvitations(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]InvitationRecord, error)
}

// InvitationRecord is a roster member as ListInvitations reads them.
type InvitationRecord struct {
	MemberID   shared.MemberID
	Name       string
	Invited    bool
	Attendance *participation.AttendanceStatus // nil when not answered
}

// EntryQueryService reads the entries of a race together with the names of their members.
// It is implemented in the infrastructure layer.
type EntryQueryService interface {
	// ListEntries returns the entries of the race. It returns an error wrapping
	// shared.ErrNotFound when the team has no such race.
	ListEntries(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]EntryRecord, error)
}

// EntryRecord is an entry as ListEntries reads it.
type EntryRecord struct {
	ID      shared.EntryID
	Name    string
	Members []EntryMemberRecord
}

// EntryMemberRecord is a member of an entry as ListEntries reads them.
type EntryMemberRecord struct {
	MemberID shared.MemberID
	Name     string
}
