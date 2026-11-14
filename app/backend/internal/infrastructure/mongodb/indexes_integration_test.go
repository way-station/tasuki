//go:build integration

package mongodb

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEnsureIndexes(t *testing.T) {
	client := newLapRecordsClient(t)
	ctx := context.Background()

	t.Run("can run again", func(t *testing.T) {
		if err := EnsureIndexes(ctx, client); err != nil {
			t.Fatalf("EnsureIndexes again: %v", err)
		}
	})

	t.Run("keeps one document per race and member", func(t *testing.T) {
		race, err := bsonUUID(lapRaceID)
		if err != nil {
			t.Fatal(err)
		}
		member, err := bsonUUID(lapTaroID)
		if err != nil {
			t.Fatal(err)
		}
		collection := client.Database().Collection(lapRecordsCollection)
		document := bson.D{{Key: "race_id", Value: race}, {Key: "roster_member_id", Value: member}, {Key: "laps", Value: bson.A{}}}
		if _, err := collection.InsertOne(ctx, document); err != nil {
			t.Fatal(err)
		}
		if _, err := collection.InsertOne(ctx, document); err == nil {
			t.Error("a second document of the same race and member was stored")
		}
	})
}
