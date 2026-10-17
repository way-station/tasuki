package roster

import "github.com/way-station/tasuki/app/backend/internal/domain/shared"

// InvitedRace is a race a member is invited to, with its race date.
type InvitedRace struct {
	RaceID   shared.RaceID
	RaceDate shared.Date
}

// LeaveService holds the rules for marking a member as left that involve their races.
type LeaveService struct{}

// InvitationsToCancel returns the races whose invitations are cancelled when the member
// leaves: races dated today or later. Past races are not changed.
func (LeaveService) InvitationsToCancel(today shared.Date, invited []InvitedRace) []shared.RaceID {
	var ids []shared.RaceID
	for _, r := range invited {
		if !r.RaceDate.Before(today) {
			ids = append(ids, r.RaceID)
		}
	}
	return ids
}
