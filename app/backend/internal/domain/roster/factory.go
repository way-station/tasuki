package roster

import (
	"strings"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MemberFactory assembles Member values.
type MemberFactory struct {
	ids shared.IDGenerator
}

// NewMemberFactory returns a MemberFactory that takes new IDs from ids.
func NewMemberFactory(ids shared.IDGenerator) MemberFactory {
	return MemberFactory{ids: ids}
}

// New returns a new roster member of team. The name is required.
func (f MemberFactory) New(teamID shared.TeamID, name string) (*Member, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.NewValidationError("name", "is required")
	}
	return &Member{id: shared.MemberID(f.ids.NewID()), teamID: teamID, name: name}, nil
}

// Reconstruct returns a member read from storage. Repositories use it; the values are
// trusted because they were validated when they were stored.
func (f MemberFactory) Reconstruct(id shared.MemberID, teamID shared.TeamID, name string, leftAt *time.Time) *Member {
	return &Member{id: id, teamID: teamID, name: name, leftAt: leftAt}
}
