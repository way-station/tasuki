package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LapRecordReader reads the laps of lap_records for the lap query service, without loading
// entities (ADR-0011).
type LapRecordReader struct {
	collection *mongo.Collection
}

// NewLapRecordReader returns a LapRecordReader that uses client.
func NewLapRecordReader(client *Client) *LapRecordReader {
	return &LapRecordReader{collection: client.Database().Collection(lapRecordsCollection)}
}

// RaceLaps returns the laps of each roster member who has laps for the race, in submission
// order.
func (reader *LapRecordReader) RaceLaps(ctx context.Context, raceID shared.RaceID) (map[shared.MemberID][]laps.LapRecord, error) {
	race, err := bsonUUID(string(raceID))
	if err != nil {
		return nil, err
	}
	cursor, err := reader.collection.Find(ctx, bson.D{{Key: "race_id", Value: race}})
	if err != nil {
		return nil, fmt.Errorf("read the laps of race %s: %w", raceID, err)
	}
	var documents []lapRecordDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("read the laps of race %s: %w", raceID, err)
	}
	byMember := make(map[shared.MemberID][]laps.LapRecord, len(documents))
	for _, document := range documents {
		memberID, err := idOf(document.RosterMemberID)
		if err != nil {
			return nil, err
		}
		records, err := lapRecordsOf(document.Laps)
		if err != nil {
			return nil, err
		}
		byMember[shared.MemberID(memberID)] = records
	}
	return byMember, nil
}

// MemberLaps returns the member's laps for the race in submission order, or none.
func (reader *LapRecordReader) MemberLaps(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) ([]laps.LapRecord, error) {
	filter, err := recordFilter(raceID, memberID)
	if err != nil {
		return nil, err
	}
	var document lapRecordDocument
	err = reader.collection.FindOne(ctx, filter).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return []laps.LapRecord{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read the laps of %s in race %s: %w", memberID, raceID, err)
	}
	return lapRecordsOf(document.Laps)
}

func lapRecordsOf(documents []lapDocument) ([]laps.LapRecord, error) {
	records := make([]laps.LapRecord, 0, len(documents))
	for _, document := range documents {
		lapID, err := idOf(document.LapID)
		if err != nil {
			return nil, err
		}
		lapTime, err := shared.PaceFromSeconds(document.Seconds)
		if err != nil {
			return nil, fmt.Errorf("lap %s has an invalid stored time: %w", lapID, err)
		}
		records = append(records, laps.LapRecord{ID: shared.LapID(lapID), Time: lapTime})
	}
	return records, nil
}
