//go:build integration

package mongodb

import (
	"context"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestLapRecordReader(t *testing.T) {
	client := newLapRecordsClient(t)
	repository := newLapRecordRepository(client)
	ctx := context.Background()
	first, second := newLap(t, "3:20"), newLap(t, "3:25")
	taro := lap.RecordFactory{}.New(lapRaceID, lapTaroID)
	taro.Add(first, second)
	other := lap.RecordFactory{}.New(lapOtherRaceID, lapHanakoID)
	other.Add(newLap(t, "4:00"))
	for _, record := range []*lap.Record{taro, other} {
		if err := repository.Save(ctx, record); err != nil {
			t.Fatal(err)
		}
	}
	reader := NewLapRecordReader(client)

	t.Run("reads the laps of each member who has laps for the race", func(t *testing.T) {
		byMember, err := reader.RaceLaps(ctx, lapRaceID)
		if err != nil {
			t.Fatalf("RaceLaps: %v", err)
		}
		laps := byMember[lapTaroID]
		if len(byMember) != 1 || len(laps) != 2 || laps[0].ID != first.ID() || laps[0].Time.String() != "3:20" || laps[1].ID != second.ID() {
			t.Errorf("RaceLaps = %+v, want the two laps of Taro", byMember)
		}
	})

	t.Run("reads the member's laps for the race", func(t *testing.T) {
		laps, err := reader.MemberLaps(ctx, lapRaceID, lapTaroID)
		if err != nil || len(laps) != 2 || laps[1].Time.String() != "3:25" {
			t.Errorf("MemberLaps = %+v, %v, want the two laps", laps, err)
		}
	})

	t.Run("reads no laps of a member without a record", func(t *testing.T) {
		laps, err := reader.MemberLaps(ctx, lapRaceID, shared.MemberID(lapHanakoID))
		if err != nil || laps == nil || len(laps) != 0 {
			t.Errorf("MemberLaps = %#v, %v, want an empty list", laps, err)
		}
	})
}
