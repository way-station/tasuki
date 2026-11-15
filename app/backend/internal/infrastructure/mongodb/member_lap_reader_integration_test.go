//go:build integration

package mongodb

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestMemberLapReader(t *testing.T) {
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
	collection := client.Database().Collection("lap_records")
	if err := collection.Drop(ctx); err != nil {
		t.Fatal(err)
	}

	const (
		taroID   = "01900000-0000-7000-8000-000000000e21"
		hanakoID = "01900000-0000-7000-8000-000000000e22"
		hakone   = "01900000-0000-7000-8000-000000000e41"
		empty    = "01900000-0000-7000-8000-000000000e42"
		firstLap = "01900000-0000-7000-8000-000000000e51"
		secondLp = "01900000-0000-7000-8000-000000000e52"
	)
	ids := map[string]bson.Binary{}
	for _, id := range []string{taroID, hanakoID, hakone, empty, firstLap, secondLp} {
		binary, err := bsonUUID(id)
		if err != nil {
			t.Fatal(err)
		}
		ids[id] = binary
	}
	submittedAt := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	documents := []any{
		bson.D{{Key: "race_id", Value: ids[hakone]}, {Key: "roster_member_id", Value: ids[taroID]}, {Key: "laps", Value: bson.A{
			bson.D{{Key: "lap_id", Value: ids[firstLap]}, {Key: "seconds", Value: 200}, {Key: "submitted_at", Value: submittedAt}},
			bson.D{{Key: "lap_id", Value: ids[secondLp]}, {Key: "seconds", Value: 205}, {Key: "submitted_at", Value: submittedAt}},
		}}},
		bson.D{{Key: "race_id", Value: ids[empty]}, {Key: "roster_member_id", Value: ids[taroID]}, {Key: "laps", Value: bson.A{}}},
		bson.D{{Key: "race_id", Value: ids[hakone]}, {Key: "roster_member_id", Value: ids[hanakoID]}, {Key: "laps", Value: bson.A{
			bson.D{{Key: "lap_id", Value: ids[firstLap]}, {Key: "seconds", Value: 300}, {Key: "submitted_at", Value: submittedAt}},
		}}},
	}
	if _, err := collection.InsertMany(ctx, documents); err != nil {
		t.Fatal(err)
	}

	byRace, err := NewMemberLapReader(client).LapsByRace(ctx, taroID)
	if err != nil {
		t.Fatalf("LapsByRace: %v", err)
	}
	laps := byRace[shared.RaceID(hakone)]
	if len(byRace) != 1 || len(laps) != 2 || laps[0].ID != firstLap || laps[0].Time.String() != "3:20" || laps[1].Time.String() != "3:25" {
		t.Errorf("LapsByRace = %+v, want the two laps of Taro in Hakone only", byRace)
	}
}
