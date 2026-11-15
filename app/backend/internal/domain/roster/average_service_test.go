package roster

import (
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func paces(t *testing.T, texts ...string) []shared.Pace {
	t.Helper()
	parsed := make([]shared.Pace, len(texts))
	for i, text := range texts {
		lapPace, err := shared.ParsePace(text)
		if err != nil {
			t.Fatal(err)
		}
		parsed[i] = lapPace
	}
	return parsed
}

func memberWith(t *testing.T, average string, lapCount int) *Member {
	t.Helper()
	return NewMemberFactory(fixedIDs{id: "m1"}).Reconstruct("m1", "team1", "Sato", nil, paces(t, average)[0], lapCount)
}

func TestNewMemberHasTheDefaultAverage(t *testing.T) {
	member := newMember(t)
	if member.Average().String() != "3:40" || member.LapCount() != 0 {
		t.Errorf("Average() = %s, LapCount() = %d, want 3:40 and 0", member.Average(), member.LapCount())
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
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			member := memberWith(t, testCase.average, testCase.lapCount)
			AverageService{}.AddLaps(member, paces(t, testCase.laps...))
			if member.Average().String() != testCase.want || member.LapCount() != testCase.wantCount {
				t.Errorf("average %s, count %d, want %s and %d", member.Average(), member.LapCount(), testCase.want, testCase.wantCount)
			}
		})
	}
}

func TestRecalculate(t *testing.T) {
	member := memberWith(t, "3:28", 8)
	AverageService{}.Recalculate(member, paces(t, "3:20", "3:25"))
	if member.Average().String() != "3:23" || member.LapCount() != 2 { // 202.5 → 203
		t.Errorf("average %s, count %d, want 3:23 and 2", member.Average(), member.LapCount())
	}
	AverageService{}.Recalculate(member, nil)
	if member.Average().String() != "3:40" || member.LapCount() != 0 {
		t.Errorf("with no laps: average %s, count %d, want 3:40 and 0", member.Average(), member.LapCount())
	}
}
