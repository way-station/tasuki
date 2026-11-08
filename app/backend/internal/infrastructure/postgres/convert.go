package postgres

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Conversions between the generated types and the domain's. Generated types stay in this
// package (ADR-0015).

// parseID converts a domain ID to a uuid column value.
func parseID[ID ~string](id ID) (uuid.UUID, error) {
	parsed, err := uuid.Parse(string(id))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid id %q: %w", id, err)
	}
	return parsed, nil
}

// parseOptionalID converts an optional domain ID to a nullable uuid value.
func parseOptionalID[ID ~string](id *ID) (*uuid.UUID, error) {
	if id == nil {
		return nil, nil
	}
	parsed, err := parseID(*id)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// toInt32 converts a domain integer to an integer column value.
func toInt32(value int, what string) (int32, error) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, fmt.Errorf("%s %d does not fit in an integer column", what, value)
	}
	return int32(value), nil
}

func timestamptz(at time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: at, Valid: true}
}

func timeOrNil(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	at := value.Time
	return &at
}

// notFound turns pgx.ErrNoRows into an error wrapping shared.ErrNotFound.
func notFound(err error, what string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", what, shared.ErrNotFound)
	}
	return err
}
