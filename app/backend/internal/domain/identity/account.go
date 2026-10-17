// Package identity is the domain of logging in and joining the team: accounts, invite
// links, and the links between accounts and roster members (ADR-0006, ADR-0007, ADR-0008).
package identity

import (
	"slices"
	"strings"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Account is a login identity (a LINE user). It does not belong to a team (ADR-0008).
// It is created by AccountFactory.
type Account struct {
	id          shared.AccountID
	lineUserID  string
	displayName string
}

// ID returns the account's ID.
func (a *Account) ID() shared.AccountID { return a.id }

// LineUserID returns the LINE user ID (the `sub` claim, ADR-0007).
func (a *Account) LineUserID() string { return a.lineUserID }

// DisplayName returns the display name from LINE.
func (a *Account) DisplayName() string { return a.displayName }

// UpdateDisplayName replaces the display name with the one LINE returned at the latest login.
// A blank name is ignored.
func (a *Account) UpdateDisplayName(name string) {
	if name = strings.TrimSpace(name); name != "" {
		a.displayName = name
	}
}

// AccountFactory assembles Account values.
type AccountFactory struct {
	ids shared.IDGenerator
}

// NewAccountFactory returns an AccountFactory that takes new IDs from ids.
func NewAccountFactory(ids shared.IDGenerator) AccountFactory {
	return AccountFactory{ids: ids}
}

// New returns a new account for a LINE user.
func (f AccountFactory) New(lineUserID, displayName string) (*Account, error) {
	if strings.TrimSpace(lineUserID) == "" {
		return nil, shared.NewValidationError("line_user_id", "is required")
	}
	return &Account{
		id:          shared.AccountID(f.ids.NewID()),
		lineUserID:  lineUserID,
		displayName: strings.TrimSpace(displayName),
	}, nil
}

// Reconstruct returns an account read from storage.
func (f AccountFactory) Reconstruct(id shared.AccountID, lineUserID, displayName string) *Account {
	return &Account{id: id, lineUserID: lineUserID, displayName: displayName}
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

// IsOrganizer reports whether the account is an organizer.
func (p OrganizerPolicy) IsOrganizer(a *Account) bool {
	return slices.Contains(p.lineUserIDs, a.lineUserID)
}
