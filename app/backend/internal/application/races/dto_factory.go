package races

import (
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RaceDTOFactory assembles RaceSummaryDTO and RaceDTO values. It leaves out what visible
// does not allow.
type RaceDTOFactory struct{}

// CreateSummary returns a race of the list.
func (RaceDTOFactory) CreateSummary(id shared.RaceID, name string, raceDate shared.Date, answer race.Answer, runningOrderConfirmed bool, visible race.Visible) RaceSummaryDTO {
	return RaceSummaryDTO{
		id:                    id,
		name:                  name,
		raceDate:              raceDate,
		answer:                answer,
		runningOrderConfirmed: visibleBool(runningOrderConfirmed, visible.RunningOrderSummary),
	}
}

// Create returns a race. The version is returned to organizers only.
func (RaceDTOFactory) Create(id shared.RaceID, name string, version int, description race.Description, answer race.Answer, runningOrderConfirmed, canSubmitLaps, organizer bool, visible race.Visible) RaceDTO {
	dto := RaceDTO{
		id:                    id,
		name:                  name,
		description:           DescriptionDTO{description: description, deadlineVisible: visible.Deadline},
		answer:                answer,
		runningOrderConfirmed: visibleBool(runningOrderConfirmed, visible.RunningOrderSummary),
		canSubmitLaps:         canSubmitLaps,
	}
	if organizer {
		dto.version = &version
	}
	return dto
}

func visibleBool(value, visible bool) *bool {
	if !visible {
		return nil
	}
	return &value
}
