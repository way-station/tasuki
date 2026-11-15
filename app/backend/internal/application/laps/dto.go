package laps

import (
	"slices"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MyLapsDTO is the caller's laps for a race as the my-laps endpoints return them. It is
// created by LapDTOFactory.
type MyLapsDTO struct {
	laps     []LapDTO
	average  shared.Pace
	lapCount int
}

// Laps returns the laps in submission order.
func (dto MyLapsDTO) Laps() []LapDTO { return slices.Clone(dto.laps) }

// Average returns the average time per km over every race.
func (dto MyLapsDTO) Average() shared.Pace { return dto.average }

// LapCount returns the number of laps counted in the average.
func (dto MyLapsDTO) LapCount() int { return dto.lapCount }

// LapDTO is one lap.
type LapDTO struct {
	id      shared.LapID
	lapTime shared.Pace
}

// ID returns the lap's ID.
func (dto LapDTO) ID() shared.LapID { return dto.id }

// Time returns the lap time per km.
func (dto LapDTO) Time() shared.Pace { return dto.lapTime }

// RunnerLapsDTO is a runner with their laps for a race, as organizers see them. It is
// created by LapDTOFactory.
type RunnerLapsDTO struct {
	memberID shared.MemberID
	name     string
	laps     []LapDTO
}

// MemberID returns the roster member's ID.
func (dto RunnerLapsDTO) MemberID() shared.MemberID { return dto.memberID }

// Name returns the roster member's name.
func (dto RunnerLapsDTO) Name() string { return dto.name }

// Laps returns the laps in submission order.
func (dto RunnerLapsDTO) Laps() []LapDTO { return slices.Clone(dto.laps) }

// ProfileDTO is the caller's profile. It is created by LapDTOFactory.
type ProfileDTO struct {
	name     string
	average  shared.Pace
	lapCount int
}

// Name returns the roster member's name.
func (dto ProfileDTO) Name() string { return dto.name }

// Average returns the average time per km.
func (dto ProfileDTO) Average() shared.Pace { return dto.average }

// LapCount returns the number of laps counted in the average.
func (dto ProfileDTO) LapCount() int { return dto.lapCount }
