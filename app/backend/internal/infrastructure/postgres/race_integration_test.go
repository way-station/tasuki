//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

const testTeamID shared.TeamID = "01900000-0000-7000-8000-000000000001"

func saveTestTeam(t *testing.T, client *Client) {
	t.Helper()
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Team')", string(testTeamID))
}

func testDescription(t *testing.T, params race.DescriptionParams) race.Description {
	t.Helper()
	description, err := race.NewDescription(params)
	if err != nil {
		t.Fatal(err)
	}
	return description
}

func distanceDescription(t *testing.T, legs ...float64) race.Description {
	t.Helper()
	date, err := shared.ParseDate("2026-11-23")
	if err != nil {
		t.Fatal(err)
	}
	target, err := shared.ParseTotalTime("2:45:00")
	if err != nil {
		t.Fatal(err)
	}
	return testDescription(t, race.DescriptionParams{
		Type: race.Distance, RaceDate: date, Deadline: testNow, Text: "Bring a tasuki",
		LegDistancesKm: legs, TargetTotalTime: &target,
	})
}

func timeDescription(t *testing.T) race.Description {
	t.Helper()
	date, err := shared.ParseDate("2026-12-06")
	if err != nil {
		t.Fatal(err)
	}
	laps := 30
	return testDescription(t, race.DescriptionParams{
		Type: race.Time, RaceDate: date, Deadline: testNow, TimeLimitMinutes: 180, LapDistanceKm: 2, TargetLaps: &laps,
	})
}

func storedKeys(t *testing.T, client *Client, id shared.RaceID) []string {
	t.Helper()
	var data []byte
	if err := client.Pool().QueryRow(context.Background(), "SELECT description FROM races WHERE id = $1", string(id)).Scan(&data); err != nil {
		t.Fatal(err)
	}
	var stored map[string]any
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatal(err)
	}
	return slices.Sorted(maps.Keys(stored))
}

func TestRaceRepository(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	ctx := context.Background()
	factory := race.NewFactory(random.IDGenerator{})
	repository := NewRaceRepository(client, factory)

	created, err := factory.New(testTeamID, "Ekiden", distanceDescription(t, 5, 3.2))
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, created); err != nil {
		t.Fatalf("Save new: %v", err)
	}

	found, err := repository.FindByID(ctx, created.ID())
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	description := found.Description()
	target, _ := description.TargetTotalTime()
	if found.TeamID() != testTeamID || found.Name() != "Ekiden" || found.Version() != 0 ||
		description.Type() != race.Distance || description.RaceDate().String() != "2026-11-23" || !description.Deadline().Equal(testNow) ||
		description.Text() != "Bring a tasuki" || !slices.Equal(description.LegDistancesKm(), []float64{5, 3.2}) || target.String() != "2:45:00" {
		t.Errorf("FindByID = %+v, want the saved race", found)
	}
	wantKeys := []string{"deadline", "leg_distances_km", "race_date", "target_total_seconds", "text", "type"}
	if keys := storedKeys(t, client, created.ID()); !slices.Equal(keys, wantKeys) {
		t.Errorf("stored keys = %v, want %v", keys, wantKeys)
	}

	t.Run("updates when the version is still the one read", func(t *testing.T) {
		stale, err := repository.FindByID(ctx, created.ID())
		if err != nil {
			t.Fatal(err)
		}
		current, err := repository.FindByID(ctx, created.ID())
		if err != nil {
			t.Fatal(err)
		}
		if err := current.Update("Ekiden 2026", timeDescription(t), 0); err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, current); err != nil {
			t.Fatalf("Save update: %v", err)
		}
		updated, err := repository.FindByID(ctx, created.ID())
		if err != nil {
			t.Fatal(err)
		}
		laps, _ := updated.Description().TargetLaps()
		if updated.Name() != "Ekiden 2026" || updated.Version() != 1 || updated.Description().Type() != race.Time ||
			updated.Description().TimeLimitMinutes() != 180 || updated.Description().LapDistanceKm() != 2 || laps != 30 {
			t.Errorf("FindByID after Save = %+v", updated)
		}
		wantKeys := []string{"deadline", "lap_distance_km", "race_date", "target_laps", "text", "time_limit_minutes", "type"}
		if keys := storedKeys(t, client, created.ID()); !slices.Equal(keys, wantKeys) {
			t.Errorf("stored keys = %v, want %v", keys, wantKeys)
		}

		if err := stale.Update("Stale edit", distanceDescription(t, 5), 0); err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, stale); !errors.Is(err, shared.ErrVersionConflict) {
			t.Errorf("Save stale error = %v, want ErrVersionConflict", err)
		}
	})

	if _, err := repository.FindByID(ctx, shared.RaceID(random.IDGenerator{}.NewID())); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByID(unknown) error = %v, want ErrNotFound", err)
	}
}
