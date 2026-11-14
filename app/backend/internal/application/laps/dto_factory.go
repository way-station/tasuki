package laps

// LapDTOFactory assembles the DTOs of this package.
type LapDTOFactory struct{}

// CreateMyLaps returns the member's laps of the record.
func (LapDTOFactory) CreateMyLaps(record MyLapsRecord) MyLapsDTO {
	return MyLapsDTO{laps: lapsOf(record.Laps), average: record.Average, lapCount: record.LapCount}
}

// CreateRunnerLaps returns the runner's laps of the record.
func (LapDTOFactory) CreateRunnerLaps(record RunnerLapsRecord) RunnerLapsDTO {
	return RunnerLapsDTO{memberID: record.MemberID, name: record.Name, laps: lapsOf(record.Laps)}
}

// CreateProfile returns the profile of the record.
func (LapDTOFactory) CreateProfile(record ProfileRecord) ProfileDTO {
	return ProfileDTO{name: record.Name, average: record.Average, lapCount: record.LapCount}
}

func lapsOf(records []LapRecord) []LapDTO {
	laps := make([]LapDTO, 0, len(records))
	for _, record := range records {
		laps = append(laps, LapDTO{id: record.ID, lapTime: record.Time})
	}
	return laps
}
