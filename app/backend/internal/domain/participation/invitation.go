// Package participation is the domain of who takes part in a race: invitations,
// attendance answers, and entries (docs/design/database.md).
package participation

import (
	"context"
	"errors"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Invitation means an organizer invited a roster member to a race. Cancelling it deletes it,
// and the member's attendance and entry membership go with it. It is created by
// InvitationFactory.
type Invitation struct {
	raceID   shared.RaceID
	memberID shared.MemberID
}

// RaceID returns the race.
func (i *Invitation) RaceID() shared.RaceID { return i.raceID }

// MemberID returns the invited roster member.
func (i *Invitation) MemberID() shared.MemberID { return i.memberID }

// InvitationFactory assembles Invitation values.
type InvitationFactory struct{}

// New returns an invitation of the member to the race. Members who have left cannot be invited.
func (InvitationFactory) New(raceID shared.RaceID, memberID shared.MemberID, memberLeft bool) (*Invitation, error) {
	if memberLeft {
		return nil, shared.NewValidationError("roster_member_id", "the member has left")
	}
	return &Invitation{raceID: raceID, memberID: memberID}, nil
}

// Reconstruct returns an invitation read from storage.
func (InvitationFactory) Reconstruct(raceID shared.RaceID, memberID shared.MemberID) *Invitation {
	return &Invitation{raceID: raceID, memberID: memberID}
}

// AttendanceStatus is a member's answer for a race.
type AttendanceStatus string

// Attendance answers.
const (
	Attending    AttendanceStatus = "attending"
	NotAttending AttendanceStatus = "not_attending"
	Undecided    AttendanceStatus = "undecided"
)

// ErrAnswerFinal means the member already answered "not attending", which is final.
// The API returns 404, because the race is no longer visible to the member.
var ErrAnswerFinal = errors.New(`"not attending" is final`)

// Attendance is an invited member's answer for a race. It is created by AttendanceFactory
// on the first answer.
type Attendance struct {
	raceID   shared.RaceID
	memberID shared.MemberID
	status   AttendanceStatus
}

// RaceID returns the race.
func (a *Attendance) RaceID() shared.RaceID { return a.raceID }

// MemberID returns the roster member.
func (a *Attendance) MemberID() shared.MemberID { return a.memberID }

// Status returns the answer.
func (a *Attendance) Status() AttendanceStatus { return a.status }

// Change replaces the answer. It returns leftAttending = true when the answer changes away
// from "attending": the caller then removes the member from the entry and the running order.
func (a *Attendance) Change(to AttendanceStatus) (leftAttending bool, err error) {
	if err := validStatus(to); err != nil {
		return false, err
	}
	if a.status == NotAttending {
		return false, ErrAnswerFinal
	}
	leftAttending = a.status == Attending && to != Attending
	a.status = to
	return leftAttending, nil
}

func validStatus(s AttendanceStatus) error {
	switch s {
	case Attending, NotAttending, Undecided:
		return nil
	}
	return shared.NewValidationError("status", "must be attending, not_attending, or undecided")
}

// AttendanceFactory assembles Attendance values.
type AttendanceFactory struct{}

// New returns the first answer of an invited member.
func (AttendanceFactory) New(inv *Invitation, status AttendanceStatus) (*Attendance, error) {
	if err := validStatus(status); err != nil {
		return nil, err
	}
	return &Attendance{raceID: inv.raceID, memberID: inv.memberID, status: status}, nil
}

// Reconstruct returns an answer read from storage.
func (AttendanceFactory) Reconstruct(raceID shared.RaceID, memberID shared.MemberID, status AttendanceStatus) *Attendance {
	return &Attendance{raceID: raceID, memberID: memberID, status: status}
}

// AttendancePolicy decides who may see attendance answers.
type AttendancePolicy struct{}

// CanSeeAllAnswers reports whether the viewer may see every member's answer.
// Only organizers may; members see only their own answer (ADR-0006).
func (AttendancePolicy) CanSeeAllAnswers(organizer bool) bool {
	return organizer
}

// InvitationRepository stores invitations. Errors for missing ones wrap shared.ErrNotFound.
type InvitationRepository interface {
	Find(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) (*Invitation, error)
	// ListByMember returns the member's invitations, for cancelling them when the member leaves.
	ListByMember(ctx context.Context, memberID shared.MemberID) ([]*Invitation, error)
	Save(ctx context.Context, i *Invitation) error
	// Delete cancels the invitation; the attendance and the entry membership are deleted with it.
	Delete(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error
}

// AttendanceRepository stores attendance answers. Errors for missing ones wrap shared.ErrNotFound.
type AttendanceRepository interface {
	Find(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) (*Attendance, error)
	Save(ctx context.Context, a *Attendance) error
}
