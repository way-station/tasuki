//go:build integration

package mongodb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestRunningOrderHistoryRepository(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := Connect(ctx, uri, "tasuki_test")
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer func() { _ = client.Close(ctx) }()
	collection := client.Database().Collection("running_order_history")
	if err := collection.Drop(ctx); err != nil {
		t.Fatal(err)
	}

	const (
		runningOrderID = "01900000-0000-7000-8000-0000000000c1"
		entryID        = "01900000-0000-7000-8000-0000000000e1"
		raceID         = "01900000-0000-7000-8000-0000000000a1"
		userID         = "01900000-0000-7000-8000-0000000000b1"
		memberID       = "01900000-0000-7000-8000-0000000000d1"
	)
	runner := shared.MemberID(memberID)
	runningOrder := runningorder.NewFactory(nil).Reconstruct(runningOrderID, entryID, runningorder.Confirmed, 2, []runningorder.Assignment{{Number: 1, Runner: &runner}, {Number: 2}})
	changedAt := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	if err := NewRunningOrderHistoryRepository(client).Append(ctx, runningorder.HistoryFactory{}.Snapshot(runningOrder, raceID, userID, changedAt)); err != nil {
		t.Fatalf("Append: %v", err)
	}

	var stored historyDocument
	if err := collection.FindOne(ctx, bson.M{}).Decode(&stored); err != nil {
		t.Fatalf("FindOne: %v", err)
	}
	ids := map[string]bson.Binary{"running_order_id": stored.RunningOrderID, "race_id": stored.RaceID, "entry_id": stored.EntryID, "changed_by_user_id": stored.ChangedByUserID}
	for field, want := range map[string]string{"running_order_id": runningOrderID, "race_id": raceID, "entry_id": entryID, "changed_by_user_id": userID} {
		if binary := ids[field]; binary.Subtype != bson.TypeBinaryUUID || uuidString(binary) != want {
			t.Errorf("%s = %#v, want the UUID %s", field, binary, want)
		}
	}
	if !stored.ChangedAt.Equal(changedAt) || stored.Status != "confirmed" {
		t.Errorf("changed_at = %v, status = %q, want %v and confirmed", stored.ChangedAt, stored.Status, changedAt)
	}
	if len(stored.Legs) != 2 {
		t.Fatalf("legs = %+v, want two", stored.Legs)
	}
	if first := stored.Legs[0]; first.LegNumber != 1 || first.RosterMemberID == nil || first.RosterMemberID.Subtype != bson.TypeBinaryUUID || uuidString(*first.RosterMemberID) != memberID {
		t.Errorf("leg 1 = %+v, want the runner %s", first, memberID)
	}
	if second := stored.Legs[1]; second.LegNumber != 2 || second.RosterMemberID != nil {
		t.Errorf("leg 2 = %+v, want an empty leg", second)
	}
}

func uuidString(binary bson.Binary) string {
	id, err := uuid.FromBytes(binary.Data)
	if err != nil {
		return ""
	}
	return id.String()
}
