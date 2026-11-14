package runningorders

import (
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
)

// RunningOrderDTOFactory assembles RunningOrderDTO values.
type RunningOrderDTOFactory struct{}

// Create returns the running order of the record. The version is returned to organizers
// only; the distances, to distance races only.
func (RunningOrderDTOFactory) Create(record RunningOrderRecord, description race.Description, prediction *Prediction, organizer bool) RunningOrderDTO {
	distances := description.LegDistancesKm()
	legs := make([]LegDTO, 0, len(record.Legs))
	for _, leg := range record.Legs {
		dto := LegDTO{number: leg.Number}
		if index := leg.Number - 1; description.Type() == race.Distance && index >= 0 && index < len(distances) {
			distance := distances[index]
			dto.distanceKm = &distance
		}
		if leg.Runner != nil {
			dto.runner = &RunnerDTO{memberID: leg.Runner.MemberID, name: leg.Runner.Name}
		}
		legs = append(legs, dto)
	}
	dto := RunningOrderDTO{entryID: record.EntryID, entryName: record.EntryName, status: record.Status, legs: legs}
	if prediction != nil {
		dto.prediction = &PredictionDTO{totalTime: prediction.TotalTime, laps: prediction.Laps, hasEmptyLegs: prediction.HasEmptyLegs}
	}
	if organizer {
		version := record.Version
		dto.version = &version
	}
	return dto
}
