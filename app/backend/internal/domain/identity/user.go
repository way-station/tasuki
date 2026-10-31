// Package identity is the domain of logging in and joining the team: users, invite
// links, and the links between users and roster members (ADR-0006, ADR-0007, ADR-0008).
package identity

import (
	"slices"
	"strings"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// User is a login identity (a LINE user). It does not belong to a team (ADR-0008).
// It is created by UserFactory.
type User struct {
	id          shared.UserID
	lineUserID  string
	displayName string
}

// ID returns the user's ID.
func (user *User) ID() shared.UserID { return user.id }

// LineUserID returns the LINE user ID (the `sub` claim, ADR-0007).
func (user *User) LineUserID() string { return user.lineUserID }

// DisplayName returns the display name from LINE.
func (user *User) DisplayName() string { return user.displayName }

// UpdateDisplayName replaces the display name with the one LINE returned at the latest login.
// A blank name is ignored.
func (user *User) UpdateDisplayName(name string) {
	if name = strings.TrimSpace(name); name != "" {
		user.displayName = name
	}
}

// UserFactory assembles User values.
type UserFactory struct {
	ids shared.IDGenerator
}

// NewUserFactory returns an UserFactory that takes new IDs from ids.
func NewUserFactory(ids shared.IDGenerator) UserFactory {
	return UserFactory{ids: ids}
}

// New returns a new user for a LINE user.
func (factory UserFactory) New(lineUserID, displayName string) (*User, error) {
	if strings.TrimSpace(lineUserID) == "" {
		return nil, shared.NewValidationError("line_user_id", "is required")
	}
	return &User{
		id:          shared.UserID(factory.ids.NewID()),
		lineUserID:  lineUserID,
		displayName: strings.TrimSpace(displayName),
	}, nil
}

// Reconstruct returns a user read from storage.
func (factory UserFactory) Reconstruct(id shared.UserID, lineUserID, displayName string) *User {
	return &User{id: id, lineUserID: lineUserID, displayName: displayName}
}

// OrganizerPolicy decides who is an organizer. Organizers are listed by LINE user ID in
// configuration, not in the app (ADR-0006).
type OrganizerPolicy struct {
	lineUserIDs []string
}

// NewOrganizerPolicy returns a policy for the configured organizer LINE user IDs.
func NewOrganizerPolicy(lineUserIDs []string) OrganizerPolicy {
	return OrganizerPolicy{lineUserIDs: slices.Clone(lineUserIDs)}
}

// IsOrganizer reports whether the user is an organizer.
func (policy OrganizerPolicy) IsOrganizer(user *User) bool {
	return slices.Contains(policy.lineUserIDs, user.lineUserID)
}
