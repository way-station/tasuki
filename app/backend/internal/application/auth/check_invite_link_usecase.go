package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// CheckInviteLinkUseCase checks an invite link before login (GET /invite-links/{token}).
type CheckInviteLinkUseCase struct {
	query       InviteLinkQueryService
	linkFactory identity.InviteLinkFactory
	dtoFactory  InviteLinkDTOFactory
	now         func() time.Time
}

// NewCheckInviteLinkUseCase returns a CheckInviteLinkUseCase.
func NewCheckInviteLinkUseCase(query InviteLinkQueryService, linkFactory identity.InviteLinkFactory, dtoFactory InviteLinkDTOFactory, now func() time.Time) *CheckInviteLinkUseCase {
	return &CheckInviteLinkUseCase{query: query, linkFactory: linkFactory, dtoFactory: dtoFactory, now: now}
}

// Execute returns the invite link of token. An unknown, expired, used, or replaced link is
// identity.ErrInviteLinkUnusable.
func (useCase *CheckInviteLinkUseCase) Execute(ctx context.Context, token string) (InviteLinkDTO, error) {
	record, err := useCase.query.FindByTokenHash(ctx, identity.HashToken(token))
	if errors.Is(err, shared.ErrNotFound) {
		return InviteLinkDTO{}, identity.ErrInviteLinkUnusable
	}
	if err != nil {
		return InviteLinkDTO{}, err
	}
	link := useCase.linkFactory.Reconstruct(record.ID, record.MemberID, record.TokenHash, record.ExpiresAt, record.UsedAt, record.RevokedAt)
	if !link.Usable(useCase.now()) {
		return InviteLinkDTO{}, identity.ErrInviteLinkUnusable
	}
	return useCase.dtoFactory.Create(record.TeamName), nil
}
