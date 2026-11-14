package runningorders

import (
	"slices"

	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RunningOrderDTO is a running order as the running order endpoints return it. It is
// created by RunningOrderDTOFactory.
type RunningOrderDTO struct {
	entryID    shared.EntryID
	entryName  string
	status     runningorder.Status
	version    *int
	legs       []LegDTO
	prediction *PredictionDTO
}

// EntryID returns the entry.
func (dto RunningOrderDTO) EntryID() shared.EntryID { return dto.entryID }

// EntryName returns the entry name.
func (dto RunningOrderDTO) EntryName() string { return dto.entryName }

// Status returns draft or confirmed.
func (dto RunningOrderDTO) Status() runningorder.Status { return dto.status }

// Version returns the version, and false when the caller is not an organizer.
func (dto RunningOrderDTO) Version() (int, bool) {
	if dto.version == nil {
		return 0, false
	}
	return *dto.version, true
}

// Legs returns the legs in order.
func (dto RunningOrderDTO) Legs() []LegDTO { return slices.Clone(dto.legs) }

// Prediction returns the prediction, and false when no leg has a runner.
func (dto RunningOrderDTO) Prediction() (PredictionDTO, bool) {
	if dto.prediction == nil {
		return PredictionDTO{}, false
	}
	return *dto.prediction, true
}

// LegDTO is a leg of a running order.
type LegDTO struct {
	number     int
	distanceKm *float64
	runner     *RunnerDTO
}

// Number returns the leg number, from 1.
func (dto LegDTO) Number() int { return dto.number }

// DistanceKm returns the leg distance, and false for a time race.
func (dto LegDTO) DistanceKm() (float64, bool) {
	if dto.distanceKm == nil {
		return 0, false
	}
	return *dto.distanceKm, true
}

// Runner returns the runner, and false for an empty leg.
func (dto LegDTO) Runner() (RunnerDTO, bool) {
	if dto.runner == nil {
		return RunnerDTO{}, false
	}
	return *dto.runner, true
}

// RunnerDTO is the runner of a leg. Average times are left out: members never receive other
// members' average times (docs/design/api.md).
type RunnerDTO struct {
	memberID shared.MemberID
	name     string
}

// MemberID returns the roster member's ID.
func (dto RunnerDTO) MemberID() shared.MemberID { return dto.memberID }

// Name returns the roster member's name.
func (dto RunnerDTO) Name() string { return dto.name }

// PredictionDTO is the predicted result of a running order: a total time for a distance
// race, or a number of laps for a time race.
type PredictionDTO struct {
	totalTime    *shared.TotalTime
	laps         *int
	hasEmptyLegs bool
}

// TotalTime returns the predicted total time, and false for a time race.
func (dto PredictionDTO) TotalTime() (shared.TotalTime, bool) {
	if dto.totalTime == nil {
		return shared.TotalTime{}, false
	}
	return *dto.totalTime, true
}

// Laps returns the predicted number of laps, and false for a distance race.
func (dto PredictionDTO) Laps() (int, bool) {
	if dto.laps == nil {
		return 0, false
	}
	return *dto.laps, true
}

// HasEmptyLegs reports whether some legs were left out because they have no runner.
func (dto PredictionDTO) HasEmptyLegs() bool { return dto.hasEmptyLegs }
