package laps

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// GetProfileUseCase returns the caller's profile (GET /me/profile). Only members may call
// it; the middleware checks the role.
type GetProfileUseCase struct {
	query      ProfileQueryService
	dtoFactory LapDTOFactory
}

// NewGetProfileUseCase returns a GetProfileUseCase.
func NewGetProfileUseCase(query ProfileQueryService, dtoFactory LapDTOFactory) *GetProfileUseCase {
	return &GetProfileUseCase{query: query, dtoFactory: dtoFactory}
}

// Execute returns the name, average time, and lap count of the roster member the caller
// acts as. A caller without a roster member is an error wrapping shared.ErrNotFound.
func (useCase *GetProfileUseCase) Execute(ctx context.Context, caller auth.MeDTO) (ProfileDTO, error) {
	rosterMember, ok := caller.RosterMember()
	if !ok {
		return ProfileDTO{}, fmt.Errorf("the caller has no roster member: %w", shared.ErrNotFound)
	}
	record, err := useCase.query.GetProfile(ctx, rosterMember.ID())
	if err != nil {
		return ProfileDTO{}, err
	}
	return useCase.dtoFactory.CreateProfile(record), nil
}
