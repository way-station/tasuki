package runningorders

import (
	"math"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Prediction is the predicted result of a running order: a total time for a distance
// race, or a number of laps for a time race.
type Prediction struct {
	TotalTime    *shared.TotalTime
	Laps         *int
	HasEmptyLegs bool
}

// Predictor calculates the prediction of a running order from its runners' average times
// (docs/design/database.md). It runs in the application layer because it combines every
// runner's time (ADR-0011).
type Predictor struct{}

// Predict returns the prediction, and nil when no leg has a runner. Empty legs are left
// out.
func (Predictor) Predict(description race.Description, legs []LegRecord) (*Prediction, error) {
	distances := description.LegDistancesKm()
	var paceSum, timeSum float64
	runnerLegs := 0
	for _, leg := range legs {
		if leg.Runner == nil {
			continue
		}
		runnerLegs++
		paceSum += float64(leg.Runner.AverageSeconds)
		if index := leg.Number - 1; index >= 0 && index < len(distances) {
			timeSum += float64(leg.Runner.AverageSeconds) * distances[index]
		}
	}
	if runnerLegs == 0 {
		return nil, nil
	}
	prediction := &Prediction{HasEmptyLegs: runnerLegs < len(legs)}
	if description.Type() == race.Time {
		// The mean counts each leg, so a runner of two legs counts twice.
		distanceKm := float64(description.TimeLimitMinutes()*60) / (paceSum / float64(runnerLegs))
		laps := int(math.Floor(distanceKm / description.LapDistanceKm()))
		prediction.Laps = &laps
		return prediction, nil
	}
	totalTime, err := shared.TotalTimeFromSeconds(int(math.Floor(timeSum + 0.5)))
	if err != nil {
		return nil, err
	}
	prediction.TotalTime = &totalTime
	return prediction, nil
}
