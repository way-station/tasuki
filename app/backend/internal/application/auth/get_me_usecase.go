package auth

import (
	"context"
	"errors"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// GetMeUseCase returns who the caller is and which screens they may use (GET /me).
type GetMeUseCase struct {
	query          MeQueryService
	accountFactory identity.AccountFactory
	linkFactory    identity.MemberLinkFactory
	linkStatus     identity.LinkStatusService
	organizers     identity.OrganizerPolicy
	dtoFactory     MeDTOFactory
}

// NewGetMeUseCase returns a GetMeUseCase.
func NewGetMeUseCase(query MeQueryService, accountFactory identity.AccountFactory, linkFactory identity.MemberLinkFactory, linkStatus identity.LinkStatusService, organizers identity.OrganizerPolicy, dtoFactory MeDTOFactory) *GetMeUseCase {
	return &GetMeUseCase{query: query, accountFactory: accountFactory, linkFactory: linkFactory, linkStatus: linkStatus, organizers: organizers, dtoFactory: dtoFactory}
}

// Execute returns the caller. An account that no longer exists is ErrUnauthenticated.
// The status and the organizer role are decided by the domain (ADR-0006, ADR-0011).
func (useCase *GetMeUseCase) Execute(ctx context.Context, accountID shared.AccountID) (MeDTO, error) {
	record, err := useCase.query.FindMe(ctx, accountID)
	if errors.Is(err, shared.ErrNotFound) {
		return MeDTO{}, ErrUnauthenticated
	}
	if err != nil {
		return MeDTO{}, err
	}

	account := useCase.accountFactory.Reconstruct(accountID, record.LineUserID, record.DisplayName)
	links := make([]*identity.MemberLink, 0, len(record.Links))
	var rosterMember *RosterMemberDTO
	for _, linkRecord := range record.Links {
		link := useCase.linkFactory.Reconstruct(linkRecord.ID, accountID, linkRecord.RosterMemberID, linkRecord.InviteLinkID, linkRecord.CreatedAt, linkRecord.ApprovedAt, linkRecord.RejectedAt, linkRecord.UnlinkedAt)
		links = append(links, link)
		if link.Active() {
			rosterMember = useCase.dtoFactory.CreateRosterMember(linkRecord.RosterMemberID, linkRecord.RosterMemberName)
		}
	}

	status := useCase.linkStatus.CallerStatus(links)
	return useCase.dtoFactory.Create(record.DisplayName, useCase.organizers.IsOrganizer(account), status, rosterMember), nil
}
