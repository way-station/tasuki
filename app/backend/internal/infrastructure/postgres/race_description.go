package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// raceDescriptionJSON is races.description as stored (docs/design/database.md). Keys that
// do not belong to the race type are left out.
type raceDescriptionJSON struct {
	Type               string    `json:"type"`
	RaceDate           string    `json:"race_date"`
	Deadline           time.Time `json:"deadline"`
	Text               string    `json:"text"`
	LegDistancesKm     []float64 `json:"leg_distances_km,omitempty"`
	TimeLimitMinutes   int       `json:"time_limit_minutes,omitempty"`
	LapDistanceKm      float64   `json:"lap_distance_km,omitempty"`
	TargetTotalSeconds *int      `json:"target_total_seconds,omitempty"`
	TargetLaps         *int      `json:"target_laps,omitempty"`
}

func encodeRaceDescription(description race.Description) ([]byte, error) {
	stored := raceDescriptionJSON{
		Type:             string(description.Type()),
		RaceDate:         description.RaceDate().String(),
		Deadline:         description.Deadline(),
		Text:             description.Text(),
		LegDistancesKm:   description.LegDistancesKm(),
		TimeLimitMinutes: description.TimeLimitMinutes(),
		LapDistanceKm:    description.LapDistanceKm(),
	}
	if totalTime, ok := description.TargetTotalTime(); ok {
		seconds := totalTime.Seconds()
		stored.TargetTotalSeconds = &seconds
	}
	if laps, ok := description.TargetLaps(); ok {
		stored.TargetLaps = &laps
	}
	data, err := json.Marshal(stored)
	if err != nil {
		return nil, fmt.Errorf("encode race description: %w", err)
	}
	return data, nil
}

// decodeRaceDescription returns the parts of a stored description. They are checked when
// race.NewDescription builds the description from them.
func decodeRaceDescription(data []byte) (race.DescriptionParams, error) {
	var stored raceDescriptionJSON
	if err := json.Unmarshal(data, &stored); err != nil {
		return race.DescriptionParams{}, fmt.Errorf("decode race description: %w", err)
	}
	raceDate, err := shared.ParseDate(stored.RaceDate)
	if err != nil {
		return race.DescriptionParams{}, fmt.Errorf("decode race description: %w", err)
	}
	params := race.DescriptionParams{
		Type:             race.Type(stored.Type),
		RaceDate:         raceDate,
		Deadline:         stored.Deadline,
		Text:             stored.Text,
		LegDistancesKm:   stored.LegDistancesKm,
		TimeLimitMinutes: stored.TimeLimitMinutes,
		LapDistanceKm:    stored.LapDistanceKm,
		TargetLaps:       stored.TargetLaps,
	}
	if stored.TargetTotalSeconds != nil {
		totalTime, err := shared.TotalTimeFromSeconds(*stored.TargetTotalSeconds)
		if err != nil {
			return race.DescriptionParams{}, fmt.Errorf("decode race description: %w", err)
		}
		params.TargetTotalTime = &totalTime
	}
	return params, nil
}
