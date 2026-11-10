package runningorders

import (
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const teamID shared.TeamID = "t1"

var testNow = time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)

// distanceParams is a distance race of two legs, 5 km and 3.2 km.
func distanceParams(t *testing.T) race.DescriptionParams {
	t.Helper()
	raceDate, err := shared.ParseDate("2026-12-06")
	if err != nil {
		t.Fatal(err)
	}
	return race.DescriptionParams{Type: race.Distance, RaceDate: raceDate, Deadline: testNow, LegDistancesKm: []float64{5, 3.2}}
}

// timeParams is a time race of 180 minutes with 2 km laps.
func timeParams(t *testing.T) race.DescriptionParams {
	t.Helper()
	params := distanceParams(t)
	params.Type, params.LegDistancesKm, params.TimeLimitMinutes, params.LapDistanceKm = race.Time, nil, 180, 2
	return params
}

func description(t *testing.T, params race.DescriptionParams) race.Description {
	t.Helper()
	described, err := race.NewDescription(params)
	if err != nil {
		t.Fatal(err)
	}
	return described
}

// linkedMember returns a member linked to roster member m1.
func linkedMember() auth.MeDTO {
	factory := auth.MeDTOFactory{}
	return factory.Create("u1", "Taro", false, identity.CallerApproved, factory.CreateRosterMember("m1", "Taro"))
}

// organizer returns an organizer without a link.
func organizer() auth.MeDTO {
	return auth.MeDTOFactory{}.Create("u2", "Hanako", true, identity.CallerNone, nil)
}
