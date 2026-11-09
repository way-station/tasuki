package participations

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const teamID shared.TeamID = "t1"

var errDatabase = errors.New("database is down")

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "e1" }

type transactionKey struct{}

func inTransaction(ctx context.Context) bool {
	marked, _ := ctx.Value(transactionKey{}).(bool)
	return marked
}

// teamRace returns race r1 of the team.
func teamRace(t *testing.T, team shared.TeamID) *race.Race {
	t.Helper()
	raceDate, err := shared.ParseDate("2026-12-06")
	if err != nil {
		t.Fatal(err)
	}
	description, err := race.NewDescription(race.DescriptionParams{
		Type:           race.Distance,
		RaceDate:       raceDate,
		Deadline:       time.Date(2026, 11, 30, 23, 59, 0, 0, time.UTC),
		LegDistancesKm: []float64{5, 3.2},
	})
	if err != nil {
		t.Fatal(err)
	}
	return race.NewFactory(fixedIDs{}).Reconstruct("r1", team, "Ekiden", description, 1)
}

type fakeRaceRepository struct {
	stored *race.Race
}

func (fake *fakeRaceRepository) FindByID(_ context.Context, id shared.RaceID) (*race.Race, error) {
	if fake.stored == nil || fake.stored.ID() != id {
		return nil, fmt.Errorf("race %s: %w", id, shared.ErrNotFound)
	}
	return fake.stored, nil
}

func (fake *fakeRaceRepository) Save(context.Context, *race.Race) error {
	return errors.New("races are not saved by this package")
}

type fakeMemberRepository struct {
	stored map[shared.MemberID]*roster.Member
}

func (fake *fakeMemberRepository) FindByID(_ context.Context, id shared.MemberID) (*roster.Member, error) {
	member, ok := fake.stored[id]
	if !ok {
		return nil, fmt.Errorf("roster member %s: %w", id, shared.ErrNotFound)
	}
	return member, nil
}

func (fake *fakeMemberRepository) Save(context.Context, *roster.Member) error {
	return errors.New("roster members are not saved by this package")
}

type invitationKey struct {
	raceID   shared.RaceID
	memberID shared.MemberID
}

type fakeInvitationRepository struct {
	stored              map[invitationKey]*participation.Invitation
	findErr             error
	saved               *participation.Invitation
	deleted             *invitationKey
	deleteInTransaction bool
}

func invitedTo(raceID shared.RaceID, memberIDs ...shared.MemberID) *fakeInvitationRepository {
	fake := &fakeInvitationRepository{stored: map[invitationKey]*participation.Invitation{}}
	for _, memberID := range memberIDs {
		fake.stored[invitationKey{raceID, memberID}] = participation.InvitationFactory{}.Reconstruct(raceID, memberID)
	}
	return fake
}

func (fake *fakeInvitationRepository) Find(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) (*participation.Invitation, error) {
	if fake.findErr != nil {
		return nil, fake.findErr
	}
	invitation, ok := fake.stored[invitationKey{raceID, memberID}]
	if !ok {
		return nil, fmt.Errorf("invitation of %s to %s: %w", memberID, raceID, shared.ErrNotFound)
	}
	return invitation, nil
}

func (fake *fakeInvitationRepository) ListByMember(context.Context, shared.MemberID) ([]*participation.Invitation, error) {
	return nil, errors.New("not used by this package")
}

func (fake *fakeInvitationRepository) Save(_ context.Context, invitation *participation.Invitation) error {
	fake.saved = invitation
	return nil
}

func (fake *fakeInvitationRepository) Delete(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error {
	fake.deleted = &invitationKey{raceID, memberID}
	fake.deleteInTransaction = inTransaction(ctx)
	return nil
}

// member returns a caller linked to roster member m1.
func member() auth.MeDTO {
	factory := auth.MeDTOFactory{}
	return factory.Create("u1", "Taro", false, identity.CallerApproved, factory.CreateRosterMember("m1", "Taro"))
}

// organizer returns an organizer without a link.
func organizer() auth.MeDTO {
	return auth.MeDTOFactory{}.Create("u2", "Hanako", true, identity.CallerNone, nil)
}

func isValidation(err error) bool {
	var validationErr *shared.ValidationError
	return errors.As(err, &validationErr)
}
