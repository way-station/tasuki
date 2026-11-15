package races

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RaceSummaryDTO is a race as GET /races lists it. It is created by RaceDTOFactory.
type RaceSummaryDTO struct {
	id                    shared.RaceID
	name                  string
	raceDate              shared.Date
	answer                race.Answer
	runningOrderConfirmed *bool
}

// ID returns the race's ID.
func (dto RaceSummaryDTO) ID() shared.RaceID { return dto.id }

// Name returns the race name.
func (dto RaceSummaryDTO) Name() string { return dto.name }

// RaceDate returns the race date.
func (dto RaceSummaryDTO) RaceDate() shared.Date { return dto.raceDate }

// MyAttendance returns the caller's answer, and false when the caller is not invited.
func (dto RaceSummaryDTO) MyAttendance() (race.Answer, bool) {
	return dto.answer, dto.answer != race.NotInvited
}

// RunningOrderConfirmed returns whether the running order is confirmed, and false when
// the caller may not see it.
func (dto RaceSummaryDTO) RunningOrderConfirmed() (confirmed, visible bool) {
	if dto.runningOrderConfirmed == nil {
		return false, false
	}
	return *dto.runningOrderConfirmed, true
}

// RaceDTO is a race as GET /races/{race_id} returns it. It is created by RaceDTOFactory.
type RaceDTO struct {
	id                    shared.RaceID
	name                  string
	version               *int
	description           DescriptionDTO
	answer                race.Answer
	runningOrderConfirmed *bool
	canSubmitLaps         bool
}

// ID returns the race's ID.
func (dto RaceDTO) ID() shared.RaceID { return dto.id }

// Name returns the race name.
func (dto RaceDTO) Name() string { return dto.name }

// Version returns the version, and false when the caller is not an organizer.
func (dto RaceDTO) Version() (int, bool) {
	if dto.version == nil {
		return 0, false
	}
	return *dto.version, true
}

// Description returns the race description.
func (dto RaceDTO) Description() DescriptionDTO { return dto.description }

// MyAttendance returns the caller's answer, and false when the caller is not invited.
func (dto RaceDTO) MyAttendance() (race.Answer, bool) {
	return dto.answer, dto.answer != race.NotInvited
}

// RunningOrderConfirmed returns whether the running order is confirmed, and false when
// the caller may not see it.
func (dto RaceDTO) RunningOrderConfirmed() (confirmed, visible bool) {
	if dto.runningOrderConfirmed == nil {
		return false, false
	}
	return *dto.runningOrderConfirmed, true
}

// CanSubmitLaps reports whether the caller is a runner in a confirmed running order.
func (dto RaceDTO) CanSubmitLaps() bool { return dto.canSubmitLaps }

// DescriptionDTO is a race description as the caller may see it.
type DescriptionDTO struct {
	description     race.Description
	deadlineVisible bool
}

// Type returns the race type.
func (dto DescriptionDTO) Type() race.Type { return dto.description.Type() }

// RaceDate returns the race date.
func (dto DescriptionDTO) RaceDate() shared.Date { return dto.description.RaceDate() }

// Deadline returns the deadline, and false when the caller may not see it (ADR-0006).
func (dto DescriptionDTO) Deadline() (time.Time, bool) {
	if !dto.deadlineVisible {
		return time.Time{}, false
	}
	return dto.description.Deadline(), true
}

// Text returns the free text.
func (dto DescriptionDTO) Text() string { return dto.description.Text() }

// LegDistancesKm returns the distance of each leg (distance races).
func (dto DescriptionDTO) LegDistancesKm() []float64 { return dto.description.LegDistancesKm() }

// TimeLimitMinutes returns the time limit (time races).
func (dto DescriptionDTO) TimeLimitMinutes() int { return dto.description.TimeLimitMinutes() }

// LapDistanceKm returns the distance of one lap (time races).
func (dto DescriptionDTO) LapDistanceKm() float64 { return dto.description.LapDistanceKm() }

// TargetTotalTime returns the target total time of a distance race, if set.
func (dto DescriptionDTO) TargetTotalTime() (shared.TotalTime, bool) {
	return dto.description.TargetTotalTime()
}

// TargetLaps returns the target laps of a time race, if set.
func (dto DescriptionDTO) TargetLaps() (int, bool) { return dto.description.TargetLaps() }
