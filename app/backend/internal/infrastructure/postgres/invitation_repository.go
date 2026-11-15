package postgres

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// InvitationRepository implements participation.InvitationRepository with PostgreSQL. It
// runs in the transaction of a unit of work when there is one.
type InvitationRepository struct {
	queries *sqlcgen.Queries
	factory participation.InvitationFactory
}

var _ participation.InvitationRepository = (*InvitationRepository)(nil)

// NewInvitationRepository returns an InvitationRepository that uses client.
func NewInvitationRepository(client *Client, factory participation.InvitationFactory) *InvitationRepository {
	return &InvitationRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// Find returns the invitation of the member to the race.
func (repository *InvitationRepository) Find(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) (*participation.Invitation, error) {
	race, err := parseID(raceID)
	if err != nil {
		return nil, err
	}
	member, err := parseID(memberID)
	if err != nil {
		return nil, err
	}
	row, err := queriesFor(ctx, repository.queries).GetInvitation(ctx, sqlcgen.GetInvitationParams{RaceID: race, RosterMemberID: member})
	if err != nil {
		return nil, notFound(err, "invitation")
	}
	return repository.factory.Reconstruct(shared.RaceID(row.RaceID.String()), shared.MemberID(row.RosterMemberID.String())), nil
}

// ListByMember returns the member's invitations, oldest first.
func (repository *InvitationRepository) ListByMember(ctx context.Context, memberID shared.MemberID) ([]*participation.Invitation, error) {
	member, err := parseID(memberID)
	if err != nil {
		return nil, err
	}
	rows, err := queriesFor(ctx, repository.queries).ListInvitationsByMember(ctx, member)
	if err != nil {
		return nil, err
	}
	invitations := make([]*participation.Invitation, 0, len(rows))
	for _, row := range rows {
		invitations = append(invitations, repository.factory.Reconstruct(shared.RaceID(row.RaceID.String()), shared.MemberID(row.RosterMemberID.String())))
	}
	return invitations, nil
}

// Save stores the invitation. Saving a stored invitation does nothing.
func (repository *InvitationRepository) Save(ctx context.Context, invitation *participation.Invitation) error {
	race, err := parseID(invitation.RaceID())
	if err != nil {
		return err
	}
	member, err := parseID(invitation.MemberID())
	if err != nil {
		return err
	}
	return queriesFor(ctx, repository.queries).SaveInvitation(ctx, sqlcgen.SaveInvitationParams{RaceID: race, RosterMemberID: member})
}

// Delete cancels the invitation. The answer and the entry membership are deleted with it,
// and the member's legs of the running orders are emptied (docs/design/database.md).
func (repository *InvitationRepository) Delete(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error {
	race, err := parseID(raceID)
	if err != nil {
		return err
	}
	member, err := parseID(memberID)
	if err != nil {
		return err
	}
	return queriesFor(ctx, repository.queries).DeleteInvitation(ctx, sqlcgen.DeleteInvitationParams{RaceID: race, RosterMemberID: member})
}
