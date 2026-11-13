// Package laps holds the use cases of lap times and the profile, and the outbox steps that
// write the laps and the average times (docs/design/api.md, ADR-0011, ADR-0016).
package laps

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// The kinds of the outbox steps of this package. The Infrastructure layer registers each
// kind with the outbox and runs it with LapStepRunner (docs/design/api.md).
const (
	AddLapsKind            = "lap.add_laps"
	EditLapKind            = "lap.edit_lap"
	DeleteLapKind          = "lap.delete_lap"
	RecalculateAverageKind = "lap.recalculate_average"
	RemoveRaceLapsKind     = "lap.remove_race_laps"
)

// AddLapsStep adds newly submitted laps to the member's lap record for the race (MongoDB).
type AddLapsStep struct {
	RaceID   shared.RaceID   `json:"race_id"`
	MemberID shared.MemberID `json:"member_id"`
	Laps     []StepLap       `json:"laps"`
}

// StepLap is a lap as a step stores it.
type StepLap struct {
	ID          shared.LapID `json:"id"`
	Seconds     int          `json:"seconds"`
	SubmittedAt time.Time    `json:"submitted_at"`
}

// Kind returns AddLapsKind.
func (AddLapsStep) Kind() string { return AddLapsKind }

// EditLapStep sets the time of one of the member's laps for the race (MongoDB).
type EditLapStep struct {
	RaceID   shared.RaceID   `json:"race_id"`
	MemberID shared.MemberID `json:"member_id"`
	LapID    shared.LapID    `json:"lap_id"`
	Seconds  int             `json:"seconds"`
}

// Kind returns EditLapKind.
func (EditLapStep) Kind() string { return EditLapKind }

// DeleteLapStep deletes one of the member's laps for the race (MongoDB).
type DeleteLapStep struct {
	RaceID   shared.RaceID   `json:"race_id"`
	MemberID shared.MemberID `json:"member_id"`
	LapID    shared.LapID    `json:"lap_id"`
}

// Kind returns DeleteLapKind.
func (DeleteLapStep) Kind() string { return DeleteLapKind }

// RecalculateAverageStep recalculates the member's average time and lap count from all of
// their laps (PostgreSQL). It is the final step of an entry.
type RecalculateAverageStep struct {
	MemberID shared.MemberID `json:"member_id"`
}

// Kind returns RecalculateAverageKind.
func (RecalculateAverageStep) Kind() string { return RecalculateAverageKind }

// RemoveRaceLapsStep deletes all of the member's laps for the race (MongoDB), after they
// were taken off a confirmed running order. RecalculateAverageStep follows it.
type RemoveRaceLapsStep struct {
	RaceID   shared.RaceID   `json:"race_id"`
	MemberID shared.MemberID `json:"member_id"`
}

// Kind returns RemoveRaceLapsKind.
func (RemoveRaceLapsStep) Kind() string { return RemoveRaceLapsKind }

var (
	_ outbox.Step = AddLapsStep{}
	_ outbox.Step = EditLapStep{}
	_ outbox.Step = DeleteLapStep{}
	_ outbox.Step = RecalculateAverageStep{}
	_ outbox.Step = RemoveRaceLapsStep{}
)
