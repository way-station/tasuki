package lap

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type seqIDs struct{ count int }

func (ids *seqIDs) NewID() string { ids.count++; return "lap" + strconv.Itoa(ids.count) }

func pace(t *testing.T, text string) shared.Pace {
	t.Helper()
	parsed, err := shared.ParsePace(text)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func strings(paces []shared.Pace) []string {
	texts := make([]string, len(paces))
	for i, lapPace := range paces {
		texts[i] = lapPace.String()
	}
	return texts
}

func TestRecord(t *testing.T) {
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	laps := NewFactory(&seqIDs{})
	record := RecordFactory{}.New("r1", "m1")
	record.Add(laps.New(pace(t, "3:20"), now), laps.New(pace(t, "3:25"), now), laps.New(pace(t, "3:30"), now))

	if err := record.Edit("lap2", pace(t, "3:15")); err != nil {
		t.Fatal(err)
	}
	if err := record.Delete("lap1"); err != nil {
		t.Fatal(err)
	}
	if got := strings(record.Times()); len(got) != 2 || got[0] != "3:15" || got[1] != "3:30" {
		t.Errorf("Times() = %policy, want [3:15 3:30]", got)
	}
	if err := record.Edit("missing", pace(t, "3:00")); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Edit(missing) = %policy, want ErrNotFound", err)
	}
	if err := record.Delete("missing"); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Delete(missing) = %policy, want ErrNotFound", err)
	}
}

func TestPolicies(t *testing.T) {
	if err := (SubmissionPolicy{}).CanChange(false); !errors.Is(err, ErrNotRunner) {
		t.Errorf("CanChange(false) = %policy, want ErrNotRunner", err)
	}
	if err := (SubmissionPolicy{}).CanChange(true); err != nil {
		t.Errorf("CanChange(true) = %policy", err)
	}

	policy := VisibilityPolicy{}
	if !policy.CanSee("m1", false, "m1") || policy.CanSee("m2", false, "m1") || !policy.CanSee("", true, "m1") || policy.CanSee("", false, "") {
		t.Error("CanSee() is wrong")
	}
}
