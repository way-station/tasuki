package laps

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fakeProfileQueryService struct {
	stored ProfileRecord
	err    error
}

func (fake *fakeProfileQueryService) GetProfile(_ context.Context, memberID shared.MemberID) (ProfileRecord, error) {
	if fake.err != nil {
		return ProfileRecord{}, fake.err
	}
	if memberID != "m1" {
		return ProfileRecord{}, fmt.Errorf("member %s: %w", memberID, shared.ErrNotFound)
	}
	return fake.stored, nil
}

func TestGetProfileUseCase(t *testing.T) {
	t.Run("returns the caller's average", func(t *testing.T) {
		query := &fakeProfileQueryService{stored: ProfileRecord{Name: "Taro", Average: pace(t, "3:28"), LapCount: 8}}
		profile, err := NewGetProfileUseCase(query, LapDTOFactory{}).Execute(context.Background(), runnerCaller())
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if profile.Name() != "Taro" || profile.Average().String() != "3:28" || profile.LapCount() != 8 {
			t.Errorf("Execute() = %+v, want Taro 3:28 over 8 laps", profile)
		}
	})

	t.Run("rejects a caller without a roster member", func(t *testing.T) {
		if _, err := NewGetProfileUseCase(&fakeProfileQueryService{}, LapDTOFactory{}).Execute(context.Background(), organizer()); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute() error = %v, want shared.ErrNotFound", err)
		}
	})

	t.Run("returns the error of the query service", func(t *testing.T) {
		if _, err := NewGetProfileUseCase(&fakeProfileQueryService{err: errDatabase}, LapDTOFactory{}).Execute(context.Background(), runnerCaller()); !errors.Is(err, errDatabase) {
			t.Errorf("Execute() error = %v, want %v", err, errDatabase)
		}
	})
}
