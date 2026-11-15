package runningorders

import (
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func runner(id string, averageSeconds int) LegRecord {
	return LegRecord{Runner: &RunnerRecord{MemberID: shared.MemberID(id), Name: id, AverageSeconds: averageSeconds}}
}

func numbered(legs ...LegRecord) []LegRecord {
	for i := range legs {
		legs[i].Number = i + 1
	}
	return legs
}

func TestPredictor(t *testing.T) {
	t.Run("adds the legs of a distance race and rounds the total half up", func(t *testing.T) {
		// 215 s/km × 5 km + 220 s/km × 3.2 km = 1075 + 704 = 1779 s.
		got, err := Predictor{}.Predict(description(t, distanceParams(t)), numbered(runner("m1", 215), runner("m2", 220)))
		if err != nil || got == nil || got.TotalTime == nil || got.TotalTime.String() != "0:29:39" || got.Laps != nil || got.HasEmptyLegs {
			t.Fatalf("Predict = %+v, %v, want 0:29:39 without empty legs", got, err)
		}
		params := distanceParams(t)
		params.LegDistancesKm = []float64{5.3, 0.5}
		// 215 s/km × 5.3 km = 1139.5 s, rounded up to 1140 s; the empty leg is left out.
		if got, err := (Predictor{}).Predict(description(t, params), numbered(runner("m1", 215), LegRecord{})); err != nil || got.TotalTime.String() != "0:19:00" || !got.HasEmptyLegs {
			t.Errorf("Predict with an empty leg = %+v, %v, want 0:19:00 with empty legs", got, err)
		}
	})

	t.Run("counts each leg in the mean of a time race and rounds the laps down", func(t *testing.T) {
		// Each leg: (200 + 200 + 260) / 3 = 220 s/km, 10800 / 220 ≈ 49.1 km, 24 laps.
		// Each runner once would be 230 s/km and 23 laps.
		got, err := Predictor{}.Predict(description(t, timeParams(t)), numbered(runner("m1", 200), runner("m1", 200), runner("m2", 260)))
		if err != nil || got == nil || got.Laps == nil || *got.Laps != 24 || got.TotalTime != nil {
			t.Fatalf("Predict = %+v, %v, want 24 laps", got, err)
		}
	})

	t.Run("predicts nothing without a runner", func(t *testing.T) {
		if got, err := (Predictor{}).Predict(description(t, distanceParams(t)), numbered(LegRecord{}, LegRecord{})); err != nil || got != nil {
			t.Errorf("Predict = %+v, %v, want nil", got, err)
		}
	})
}
