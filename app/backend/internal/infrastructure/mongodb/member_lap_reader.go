package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/way-station/tasuki/app/backend/internal/application/rosters"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MemberLapReader reads a roster member's laps from lap_records (docs/design/database.md)
// for the roster query service, without loading entities (ADR-0011).
type MemberLapReader struct {
	collection *mongo.Collection
}

// NewMemberLapReader returns a MemberLapReader that uses client.
func NewMemberLapReader(client *Client) *MemberLapReader {
	return &MemberLapReader{collection: client.Database().Collection("lap_records")}
}

type memberLapRecordDocument struct {
	RaceID bson.Binary         `bson:"race_id"`
	Laps   []memberLapDocument `bson:"laps"`
}

type memberLapDocument struct {
	LapID   bson.Binary `bson:"lap_id"`
	Seconds int         `bson:"seconds"`
}

// LapsByRace returns the member's laps of each race that has some, in submission order.
func (reader *MemberLapReader) LapsByRace(ctx context.Context, memberID shared.MemberID) (map[shared.RaceID][]rosters.LapRecord, error) {
	member, err := bsonUUID(string(memberID))
	if err != nil {
		return nil, err
	}
	cursor, err := reader.collection.Find(ctx, bson.D{{Key: "roster_member_id", Value: member}})
	if err != nil {
		return nil, fmt.Errorf("read the laps of %s: %w", memberID, err)
	}
	var documents []memberLapRecordDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("read the laps of %s: %w", memberID, err)
	}
	byRace := make(map[shared.RaceID][]rosters.LapRecord, len(documents))
	for _, document := range documents {
		if len(document.Laps) == 0 {
			continue
		}
		raceID, err := idOf(document.RaceID)
		if err != nil {
			return nil, err
		}
		laps := make([]rosters.LapRecord, 0, len(document.Laps))
		for _, stored := range document.Laps {
			lapID, err := idOf(stored.LapID)
			if err != nil {
				return nil, err
			}
			lapTime, err := shared.PaceFromSeconds(stored.Seconds)
			if err != nil {
				return nil, fmt.Errorf("lap %s has an invalid stored time: %w", lapID, err)
			}
			laps = append(laps, rosters.LapRecord{ID: shared.LapID(lapID), Time: lapTime})
		}
		byRace[shared.RaceID(raceID)] = laps
	}
	return byRace, nil
}
