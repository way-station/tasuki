package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const lapRecordsCollection = "lap_records"

// LapRecordRepository implements lap.Repository with the lap_records collection
// (docs/design/database.md). There is one document per race and roster member. MongoDB does
// not join the PostgreSQL unit of work; the lap steps of the outbox write through it
// (ADR-0016).
type LapRecordRepository struct {
	collection    *mongo.Collection
	lapFactory    lap.Factory
	recordFactory lap.RecordFactory
	now           func() time.Time
}

var _ lap.Repository = (*LapRecordRepository)(nil)

// NewLapRecordRepository returns a LapRecordRepository that uses client.
func NewLapRecordRepository(client *Client, lapFactory lap.Factory, recordFactory lap.RecordFactory, now func() time.Time) *LapRecordRepository {
	return &LapRecordRepository{collection: client.Database().Collection(lapRecordsCollection), lapFactory: lapFactory, recordFactory: recordFactory, now: now}
}

type lapRecordDocument struct {
	RaceID         bson.Binary   `bson:"race_id"`
	RosterMemberID bson.Binary   `bson:"roster_member_id"`
	Laps           []lapDocument `bson:"laps"`
	UpdatedAt      time.Time     `bson:"updated_at"`
}

type lapDocument struct {
	LapID       bson.Binary `bson:"lap_id"`
	Seconds     int         `bson:"seconds"`
	SubmittedAt time.Time   `bson:"submitted_at"`
}

// Find returns the member's record for the race, or an error wrapping shared.ErrNotFound.
func (repository *LapRecordRepository) Find(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) (*lap.Record, error) {
	filter, err := recordFilter(raceID, memberID)
	if err != nil {
		return nil, err
	}
	var document lapRecordDocument
	err = repository.collection.FindOne(ctx, filter).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("lap record of %s in race %s: %w", memberID, raceID, shared.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find lap record: %w", err)
	}
	return repository.reconstruct(document)
}

// ListByMember returns the member's records for every race.
func (repository *LapRecordRepository) ListByMember(ctx context.Context, memberID shared.MemberID) ([]*lap.Record, error) {
	member, err := bsonUUID(string(memberID))
	if err != nil {
		return nil, err
	}
	cursor, err := repository.collection.Find(ctx, bson.D{{Key: "roster_member_id", Value: member}})
	if err != nil {
		return nil, fmt.Errorf("list lap records: %w", err)
	}
	var documents []lapRecordDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("list lap records: %w", err)
	}
	records := make([]*lap.Record, 0, len(documents))
	for _, document := range documents {
		record, err := repository.reconstruct(document)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// Save replaces the laps of the record, creating the document for the first laps of the
// race. Saving the same record again has the same effect.
func (repository *LapRecordRepository) Save(ctx context.Context, record *lap.Record) error {
	filter, err := recordFilter(record.RaceID(), record.MemberID())
	if err != nil {
		return err
	}
	laps := make([]lapDocument, 0, len(record.Laps()))
	for _, stored := range record.Laps() {
		lapID, err := bsonUUID(string(stored.ID()))
		if err != nil {
			return err
		}
		laps = append(laps, lapDocument{LapID: lapID, Seconds: stored.Time().Seconds(), SubmittedAt: stored.SubmittedAt()})
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "laps", Value: laps}, {Key: "updated_at", Value: repository.now()}}}}
	if _, err := repository.collection.UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("save lap record: %w", err)
	}
	return nil
}

// Delete removes the member's record for the race. Deleting a record that is gone does
// nothing.
func (repository *LapRecordRepository) Delete(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error {
	filter, err := recordFilter(raceID, memberID)
	if err != nil {
		return err
	}
	if _, err := repository.collection.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("delete lap record: %w", err)
	}
	return nil
}

func (repository *LapRecordRepository) reconstruct(document lapRecordDocument) (*lap.Record, error) {
	raceID, err := idOf(document.RaceID)
	if err != nil {
		return nil, err
	}
	memberID, err := idOf(document.RosterMemberID)
	if err != nil {
		return nil, err
	}
	laps := make([]lap.Lap, 0, len(document.Laps))
	for _, stored := range document.Laps {
		lapID, err := idOf(stored.LapID)
		if err != nil {
			return nil, err
		}
		lapTime, err := shared.PaceFromSeconds(stored.Seconds)
		if err != nil {
			return nil, fmt.Errorf("lap %s has an invalid stored time: %w", lapID, err)
		}
		laps = append(laps, repository.lapFactory.Reconstruct(shared.LapID(lapID), lapTime, stored.SubmittedAt))
	}
	return repository.recordFactory.Reconstruct(shared.RaceID(raceID), shared.MemberID(memberID), laps), nil
}

// recordFilter matches the document of the member's record for the race.
func recordFilter(raceID shared.RaceID, memberID shared.MemberID) (bson.D, error) {
	race, err := bsonUUID(string(raceID))
	if err != nil {
		return nil, err
	}
	member, err := bsonUUID(string(memberID))
	if err != nil {
		return nil, err
	}
	return bson.D{{Key: "race_id", Value: race}, {Key: "roster_member_id", Value: member}}, nil
}

// bsonUUID converts an ID to a BSON UUID (binary subtype 4).
func bsonUUID(id string) (bson.Binary, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return bson.Binary{}, fmt.Errorf("invalid id %q: %w", id, err)
	}
	return bson.Binary{Subtype: bson.TypeBinaryUUID, Data: parsed[:]}, nil
}

// idOf converts a BSON UUID back to an ID.
func idOf(binary bson.Binary) (string, error) {
	if binary.Subtype != bson.TypeBinaryUUID {
		return "", fmt.Errorf("binary subtype %d is not a UUID", binary.Subtype)
	}
	parsed, err := uuid.FromBytes(binary.Data)
	if err != nil {
		return "", fmt.Errorf("invalid stored UUID: %w", err)
	}
	return parsed.String(), nil
}
