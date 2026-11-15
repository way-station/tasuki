package laps

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// recordingOutbox keeps the entries added to it.
type recordingOutbox struct {
	entries []outbox.Entry
	err     error
}

func (fake *recordingOutbox) Add(_ context.Context, entry outbox.Entry) error {
	if fake.err != nil {
		return fake.err
	}
	fake.entries = append(fake.entries, entry)
	return nil
}

func TestRemoveLapsUseCase(t *testing.T) {
	t.Run("adds one entry for each member that deletes their laps and then recalculates their average", func(t *testing.T) {
		steps := &recordingOutbox{}
		err := NewRemoveLapsUseCase(steps, outbox.EntryFactory{}).RemoveLaps(context.Background(), "r1", []shared.MemberID{"m1", "m2"})
		if err != nil {
			t.Fatalf("RemoveLaps() error = %v", err)
		}
		want := [][]outbox.Step{
			{RemoveRaceLapsStep{RaceID: "r1", MemberID: "m1"}, RecalculateAverageStep{MemberID: "m1"}},
			{RemoveRaceLapsStep{RaceID: "r1", MemberID: "m2"}, RecalculateAverageStep{MemberID: "m2"}},
		}
		var got [][]outbox.Step
		for _, entry := range steps.entries {
			got = append(got, entry.Steps())
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("entries = %v, want %v", got, want)
		}
	})

	t.Run("adds nothing without members", func(t *testing.T) {
		steps := &recordingOutbox{}
		if err := NewRemoveLapsUseCase(steps, outbox.EntryFactory{}).RemoveLaps(context.Background(), "r1", nil); err != nil {
			t.Fatalf("RemoveLaps() error = %v", err)
		}
		if len(steps.entries) != 0 {
			t.Errorf("entries = %v, want none", steps.entries)
		}
	})

	t.Run("returns the error of the outbox", func(t *testing.T) {
		steps := &recordingOutbox{err: errDatabase}
		err := NewRemoveLapsUseCase(steps, outbox.EntryFactory{}).RemoveLaps(context.Background(), "r1", []shared.MemberID{"m1"})
		if !errors.Is(err, errDatabase) {
			t.Errorf("RemoveLaps() error = %v, want %v", err, errDatabase)
		}
	})
}
