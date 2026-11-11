//go:build integration

package mongodb

import (
	"context"
	"errors"
	"os"
	"slices"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

const (
	lapRaceID      = "01900000-0000-7000-8000-000000000a11"
	lapOtherRaceID = "01900000-0000-7000-8000-000000000a12"
	lapTaroID      = "01900000-0000-7000-8000-000000000a21"
	lapHanakoID    = "01900000-0000-7000-8000-000000000a23"
)

var lapTestNow = time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)

// newLapRecordsClient connects to the test database with an empty lap_records collection
// that has its indexes.
func newLapRecordsClient(t *testing.T) *Client {
	t.Helper()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI is not set")
	}
	ctx := context.Background()
	client, err := Connect(ctx, uri, "tasuki_test")
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(func() { _ = client.Close(ctx) })
	if err := client.Database().Collection(lapRecordsCollection).Drop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := EnsureIndexes(ctx, client); err != nil {
		t.Fatalf("EnsureIndexes: %v", err)
	}
	return client
}

func newLapRecordRepository(client *Client) *LapRecordRepository {
	return NewLapRecordRepository(client, lap.NewFactory(random.IDGenerator{}), lap.RecordFactory{}, func() time.Time { return lapTestNow })
}

// newLap returns a new lap of the time.
func newLap(t *testing.T, text string) lap.Lap {
	t.Helper()
	lapTime, err := shared.ParsePace(text)
	if err != nil {
		t.Fatal(err)
	}
	return lap.NewFactory(random.IDGenerator{}).New(lapTime, lapTestNow)
}

func timesOf(record *lap.Record) []string {
	var times []string
	for _, lapTime := range record.Times() {
		times = append(times, lapTime.String())
	}
	return times
}

func TestLapRecordRepository(t *testing.T) {
	client := newLapRecordsClient(t)
	repository := newLapRecordRepository(client)
	ctx := context.Background()

	if _, err := repository.Find(ctx, lapRaceID, lapTaroID); !errors.Is(err, shared.ErrNotFound) {
		t.Fatalf("Find before saving error = %v, want ErrNotFound", err)
	}

	first, second := newLap(t, "3:20"), newLap(t, "3:25")
	record := lap.RecordFactory{}.New(lapRaceID, lapTaroID)
	record.Add(first, second)
	for range 2 {
		if err := repository.Save(ctx, record); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	t.Run("stores one document with the laps in order", func(t *testing.T) {
		count, err := client.Database().Collection(lapRecordsCollection).CountDocuments(ctx, bson.D{})
		if err != nil || count != 1 {
			t.Fatalf("documents = %d, %v, want 1", count, err)
		}
		found, err := repository.Find(ctx, lapRaceID, lapTaroID)
		if err != nil {
			t.Fatalf("Find: %v", err)
		}
		stored := found.Laps()
		if len(stored) != 2 || stored[0].ID() != first.ID() || stored[1].ID() != second.ID() || !slices.Equal(timesOf(found), []string{"3:20", "3:25"}) || !stored[0].SubmittedAt().Equal(lapTestNow) {
			t.Errorf("Find = %v, want the two laps in order", stored)
		}
	})

	t.Run("replaces the laps of the record", func(t *testing.T) {
		found, err := repository.Find(ctx, lapRaceID, lapTaroID)
		if err != nil {
			t.Fatal(err)
		}
		if err := found.Delete(first.ID()); err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, found); err != nil {
			t.Fatalf("Save: %v", err)
		}
		if saved, err := repository.Find(ctx, lapRaceID, lapTaroID); err != nil || !slices.Equal(timesOf(saved), []string{"3:25"}) {
			t.Errorf("Find after deleting a lap = %v, %v, want 3:25", saved, err)
		}
	})

	t.Run("lists the records of the member in every race", func(t *testing.T) {
		other := lap.RecordFactory{}.New(lapOtherRaceID, lapTaroID)
		other.Add(newLap(t, "3:30"))
		hanako := lap.RecordFactory{}.New(lapRaceID, lapHanakoID)
		hanako.Add(newLap(t, "4:00"))
		for _, saved := range []*lap.Record{other, hanako} {
			if err := repository.Save(ctx, saved); err != nil {
				t.Fatal(err)
			}
		}
		records, err := repository.ListByMember(ctx, lapTaroID)
		if err != nil || len(records) != 2 {
			t.Fatalf("ListByMember = %v, %v, want two records", records, err)
		}
		for _, listed := range records {
			if listed.MemberID() != lapTaroID {
				t.Errorf("ListByMember has a record of %s", listed.MemberID())
			}
		}
	})

	t.Run("deletes the record, and deleting again does nothing", func(t *testing.T) {
		for range 2 {
			if err := repository.Delete(ctx, lapRaceID, lapTaroID); err != nil {
				t.Fatalf("Delete: %v", err)
			}
		}
		if _, err := repository.Find(ctx, lapRaceID, lapTaroID); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Find after Delete error = %v, want ErrNotFound", err)
		}
	})
}
