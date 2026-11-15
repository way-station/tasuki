// Package laps holds the use cases of lap times and the profile, and the outbox steps that
// write the laps and the average times (docs/design/api.md, ADR-0011, ADR-0016).
package laps

import (
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// SubmitLapsCommand is the laps the caller submits for a race, in the order they were
// entered. It is created by LapCommandFactory.
type SubmitLapsCommand struct {
	raceID shared.RaceID
	times  []shared.Pace
}

// EditLapCommand is the new time of one of the caller's laps. It is created by
// LapCommandFactory.
type EditLapCommand struct {
	raceID  shared.RaceID
	lapID   shared.LapID
	lapTime shared.Pace
}

// DeleteLapCommand is one of the caller's laps to delete. It is created by
// LapCommandFactory.
type DeleteLapCommand struct {
	raceID shared.RaceID
	lapID  shared.LapID
}

// LapCommandFactory assembles the commands of this package. Lap times must be MM:SS.
type LapCommandFactory struct{}

// Submit returns the command to submit the times for the race. It needs at least one time,
// and every time must be MM:SS.
func (LapCommandFactory) Submit(raceID shared.RaceID, times []string) (SubmitLapsCommand, error) {
	if len(times) == 0 {
		return SubmitLapsCommand{}, shared.NewValidationError("times", "needs at least one lap")
	}
	paces := make([]shared.Pace, 0, len(times))
	for i, text := range times {
		pace, err := shared.ParsePace(text)
		if err != nil {
			return SubmitLapsCommand{}, shared.NewValidationError(fmt.Sprintf("times[%d]", i), "must be MM:SS between 0:01 and 99:59")
		}
		paces = append(paces, pace)
	}
	return SubmitLapsCommand{raceID: raceID, times: paces}, nil
}

// Edit returns the command to set the time of the lap. The time must be MM:SS.
func (LapCommandFactory) Edit(raceID shared.RaceID, lapID shared.LapID, text string) (EditLapCommand, error) {
	pace, err := shared.ParsePace(text)
	if err != nil {
		return EditLapCommand{}, err
	}
	return EditLapCommand{raceID: raceID, lapID: lapID, lapTime: pace}, nil
}

// Delete returns the command to delete the lap.
func (LapCommandFactory) Delete(raceID shared.RaceID, lapID shared.LapID) DeleteLapCommand {
	return DeleteLapCommand{raceID: raceID, lapID: lapID}
}
