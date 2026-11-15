package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EnsureIndexes creates the indexes of docs/design/database.md. The backend calls it at
// startup; indexes that already exist are left as they are.
func EnsureIndexes(ctx context.Context, client *Client) error {
	lapRecords := []mongo.IndexModel{
		// One document per race and roster member, which Save upserts.
		{Keys: bson.D{{Key: "race_id", Value: 1}, {Key: "roster_member_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		// The laps of a member in every race.
		{Keys: bson.D{{Key: "roster_member_id", Value: 1}}},
	}
	if _, err := client.Database().Collection(lapRecordsCollection).Indexes().CreateMany(ctx, lapRecords); err != nil {
		return fmt.Errorf("create the indexes of %s: %w", lapRecordsCollection, err)
	}
	return nil
}
