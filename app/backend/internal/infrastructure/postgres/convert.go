package postgres

import (
	"errors"
	"fmt"
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
