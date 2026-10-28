package roster

import "github.com/way-station/tasuki/app/backend/internal/domain/shared"

const defaultAverageSeconds = 3*60 + 40 // 3:40

// DefaultAverage returns the average time of a member with no laps (3:40). It is not counted
// as a lap.
func DefaultAverage() shared.Pace {
	average, _ := shared.PaceFromSeconds(defaultAverageSeconds)
	return average
}

// AverageService recalculates a member's average time per km. It is a domain service:
// calculating each member's time belongs to the member's domain, and Application does not
// need to know how it is done (ADR-0011).
type AverageService struct{}

// AddLaps updates the average when new laps are submitted, from the stored average and lap
// count (docs/design/database.md):
//
//	new average = (average × lap count + sum of new laps) ÷ (lap count + number of new laps)
//
// The default average is not counted, so the first laps give their own mean.
func (AverageService) AddLaps(member *Member, laps []shared.Pace) {
	if len(laps) == 0 {
		return
	}
	total := member.average.Seconds() * member.lapCount
	for _, lapTime := range laps {
		total += lapTime.Seconds()
	}
	count := member.lapCount + len(laps)
	member.average = roundedMean(total, count)
	member.lapCount = count
}

// Recalculate sets the average from all of the member's remaining laps, after a lap is
// edited or deleted. With no laps left, the average goes back to the default.
func (AverageService) Recalculate(member *Member, all []shared.Pace) {
	if len(all) == 0 {
		member.average = DefaultAverage()
		member.lapCount = 0
		return
	}
	total := 0
	for _, lapTime := range all {
		total += lapTime.Seconds()
	}
	member.average = roundedMean(total, len(all))
	member.lapCount = len(all)
}

// roundedMean returns total ÷ count, rounded half up to whole seconds.
func roundedMean(total, count int) shared.Pace {
	mean, _ := shared.PaceFromSeconds((2*total + count) / (2 * count))
	return mean
}
