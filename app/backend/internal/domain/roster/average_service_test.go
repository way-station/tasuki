package roster

import (
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func paces(t *testing.T, ss ...string) []shared.Pace {
	t.Helper()
	ps := make([]shared.Pace, len(ss))
	for i, s := range ss {
		p, err := shared.ParsePace(s)
		if err != nil {
			t.Fatal(err)
		}
		ps[i] = p
	}
	return ps
}

func memberWith(t *testing.T, average string, lapCount int) *Member {
	t.Helper()
	return NewMemberFactory(fixedIDs{id: "m1"}).Reconstruct("m1", "team1", "Sato", nil, paces(t, average)[0], lapCount)
}

func TestNewMemberHasTheDefaultAverage(t *testing.T) {
	m := newMember(t)
	if m.Average().String() != "3:40" || m.LapCount() != 0 {
		t.Errorf("Average() = %s, LapCount() = %d, want 3:40 and 0", m.Average(), m.LapCount())
	}
}

func TestAddLaps(t *testing.T) {
	tests := []struct {
		name      string
		average   string
		lapCount  int
		laps      []string
		want      string
		wantCount int
	}{
		// The example in docs/design/database.md: 3:27.5 rounds up to 3:28.
		{name: "documented example", average: "3:30", lapCount: 5, laps: []string{"3:20", "3:25", "3:25"}, want: "3:28", wantCount: 8},
		{name: "first lap replaces the default", average: "3:40", lapCount: 0, laps: []string{"3:10"}, want: "3:10", wantCount: 1},
		{name: "first laps give their mean", average: "3:40", lapCount: 0, laps: []string{"3:10", "3:21"}, want: "3:16", wantCount: 2},
		{name: "no laps", average: "3:30", lapCount: 5, want: "3:30", wantCount: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := memberWith(t, tt.average, tt.lapCount)
			AverageService{}.AddLaps(m, paces(t, tt.laps...))
			if m.Average().String() != tt.want || m.LapCount() != tt.wantCount {
				t.Errorf("average %s, count %d, want %s and %d", m.Average(), m.LapCount(), tt.want, tt.wantCount)
			}
		})
	}
}

func TestRecalculate(t *testing.T) {
	m := memberWith(t, "3:28", 8)
	AverageService{}.Recalculate(m, paces(t, "3:20", "3:25"))
	if m.Average().String() != "3:23" || m.LapCount() != 2 { // 202.5 → 203
		t.Errorf("average %s, count %d, want 3:23 and 2", m.Average(), m.LapCount())
	}
	AverageService{}.Recalculate(m, nil)
	if m.Average().String() != "3:40" || m.LapCount() != 0 {
		t.Errorf("with no laps: average %s, count %d, want 3:40 and 0", m.Average(), m.LapCount())
	}
}
