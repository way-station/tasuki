package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fakeInviteLinkQueryService struct {
	byHash map[string]InviteLinkRecord
	byID   map[shared.InviteLinkID]InviteLinkRecord
}

func (fake fakeInviteLinkQueryService) FindByTokenHash(_ context.Context, hash []byte) (InviteLinkRecord, error) {
	if record, ok := fake.byHash[string(hash)]; ok {
		return record, nil
	}
	return InviteLinkRecord{}, shared.ErrNotFound
}

func (fake fakeInviteLinkQueryService) FindByID(_ context.Context, id shared.InviteLinkID) (InviteLinkRecord, error) {
	if record, ok := fake.byID[id]; ok {
		return record, nil
	}
	return InviteLinkRecord{}, shared.ErrNotFound
}

// inviteRecord returns the record of link, of the team named "Tasuki AC".
func inviteRecord(link *identity.InviteLink) InviteLinkRecord {
	record := InviteLinkRecord{ID: link.ID(), MemberID: link.MemberID(), TokenHash: link.TokenHash(), ExpiresAt: link.ExpiresAt(), TeamName: "Tasuki AC"}
	if usedAt, ok := link.UsedAt(); ok {
		record.UsedAt = &usedAt
	}
	if revokedAt, ok := link.RevokedAt(); ok {
		record.RevokedAt = &revokedAt
	}
	return record
}

func newInviteLink(t *testing.T) *identity.InviteLink {
	t.Helper()
	link, _, err := identity.NewInviteLinkFactory(&sequentialIDs{}, fixedToken("invite-token")).New("m1", now)
	if err != nil {
		t.Fatal(err)
	}
	return link
}

func newCheckInviteLinkUseCase(records ...InviteLinkRecord) *CheckInviteLinkUseCase {
	query := fakeInviteLinkQueryService{byHash: map[string]InviteLinkRecord{}}
	for _, record := range records {
		query.byHash[string(record.TokenHash)] = record
	}
	return NewCheckInviteLinkUseCase(query, identity.NewInviteLinkFactory(&sequentialIDs{}, fixedToken("unused")), InviteLinkDTOFactory{}, clock)
}

func TestCheckInviteLinkUseCase(t *testing.T) {
	t.Run("returns the team name of a usable link", func(t *testing.T) {
		got, err := newCheckInviteLinkUseCase(inviteRecord(newInviteLink(t))).Execute(context.Background(), "invite-token")
		if err != nil || got.TeamName() != "Tasuki AC" {
			t.Fatalf("Execute() = %q, %v, want Tasuki AC", got.TeamName(), err)
		}
	})

	used := inviteRecord(newInviteLink(t))
	used.UsedAt = timeAt(-1)
	replaced := inviteRecord(newInviteLink(t))
	replaced.RevokedAt = timeAt(-1)
	expired := inviteRecord(newInviteLink(t))
	expired.ExpiresAt = now
	tests := []struct {
		name   string
		record *InviteLinkRecord
	}{
		{name: "unknown token"},
		{name: "used", record: &used},
		{name: "replaced", record: &replaced},
		{name: "expired", record: &expired},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var records []InviteLinkRecord
			if testCase.record != nil {
				records = append(records, *testCase.record)
			}
			_, err := newCheckInviteLinkUseCase(records...).Execute(context.Background(), "invite-token")
			if !errors.Is(err, identity.ErrInviteLinkUnusable) {
				t.Errorf("Execute() error = %v, want ErrInviteLinkUnusable", err)
			}
		})
	}
}
