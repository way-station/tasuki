package participations

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ListEntriesUseCase lists the entries of a race with their members
// (GET /races/{race_id}/entries). Only organizers may call it; the middleware checks the
// role.
type ListEntriesUseCase struct {
	query      EntryQueryService
	dtoFactory ParticipationDTOFactory
	teamID     shared.TeamID
}

// NewListEntriesUseCase returns a ListEntriesUseCase for the team.
func NewListEntriesUseCase(query EntryQueryService, dtoFactory ParticipationDTOFactory, teamID shared.TeamID) *ListEntriesUseCase {
	return &ListEntriesUseCase{query: query, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns the entries. A race of another team is shared.ErrNotFound.
func (useCase *ListEntriesUseCase) Execute(ctx context.Context, raceID shared.RaceID) ([]EntryDTO, error) {
	records, err := useCase.query.ListEntries(ctx, useCase.teamID, raceID)
	if err != nil {
		return nil, err
	}
	entries := make([]EntryDTO, 0, len(records))
	for _, record := range records {
		members := make([]EntryMemberDTO, 0, len(record.Members))
		for _, member := range record.Members {
			members = append(members, useCase.dtoFactory.CreateEntryMember(member.MemberID, member.Name))
		}
		entries = append(entries, useCase.dtoFactory.CreateEntry(record.ID, record.Name, members))
	}
	return entries, nil
}
