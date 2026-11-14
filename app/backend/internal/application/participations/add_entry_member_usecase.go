package participations

import (
	"context"
	"errors"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// AddEntryMemberUseCase adds a roster member to an entry
// (PUT /races/{race_id}/entries/{entry_id}/members/{roster_member_id}). Only organizers may
// call it; the middleware checks the role.
type AddEntryMemberUseCase struct {
	races       race.Repository
	entries     participation.EntryRepository
	attendances participation.AttendanceRepository
	teamID      shared.TeamID
}

// NewAddEntryMemberUseCase returns an AddEntryMemberUseCase for the team.
func NewAddEntryMemberUseCase(races race.Repository, entries participation.EntryRepository, attendances participation.AttendanceRepository, teamID shared.TeamID) *AddEntryMemberUseCase {
	return &AddEntryMemberUseCase{races: races, entries: entries, attendances: attendances, teamID: teamID}
}

// Execute adds the member. A member whose answer is not "attending", including one who is
// not invited, is a shared.ValidationError. Adding a member who is in the entry does
// nothing. A race of another team, or an entry of another race, is shared.ErrNotFound.
func (useCase *AddEntryMemberUseCase) Execute(ctx context.Context, command EntryMemberCommand) error {
	entry, err := findRaceEntry(ctx, useCase.races, useCase.entries, useCase.teamID, command)
	if err != nil {
		return err
	}
	var answer participation.AttendanceStatus
	attendance, err := useCase.attendances.Find(ctx, command.raceID, command.memberID)
	switch {
	case err == nil:
		answer = attendance.Status()
	case !errors.Is(err, shared.ErrNotFound):
		return err
	}
	if err := entry.Add(command.memberID, answer); err != nil {
		return err
	}
	return useCase.entries.Save(ctx, entry)
}
