package rosters

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ListPendingLinksUseCase lists the links that wait for approval (GET /member-links). Only
// organizers may call it; the middleware checks the role.
type ListPendingLinksUseCase struct {
	query      RosterQueryService
	dtoFactory RosterDTOFactory
	teamID     shared.TeamID
}

// NewListPendingLinksUseCase returns a ListPendingLinksUseCase for the team.
func NewListPendingLinksUseCase(query RosterQueryService, dtoFactory RosterDTOFactory, teamID shared.TeamID) *ListPendingLinksUseCase {
	return &ListPendingLinksUseCase{query: query, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns the pending links, oldest first.
func (useCase *ListPendingLinksUseCase) Execute(ctx context.Context) ([]PendingLinkDTO, error) {
	records, err := useCase.query.ListPendingLinks(ctx, useCase.teamID)
	if err != nil {
		return nil, err
	}
	links := make([]PendingLinkDTO, 0, len(records))
	for _, record := range records {
		links = append(links, useCase.dtoFactory.CreatePendingLink(record))
	}
	return links, nil
}
