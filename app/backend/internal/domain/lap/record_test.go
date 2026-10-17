package lap

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return "lap" + strconv.Itoa(s.n) }

func pace(t *testing.T, s string) shared.Pace {
	t.Helper()
	p, err := shared.ParsePace(s)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func strings(ps []shared.Pace) []string {
	ss := make([]string, len(ps))
	for i, p := range ps {
		ss[i] = p.String()
	}
	return ss
}

func TestRecord(t *testing.T) {
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	laps := NewFactory(&seqIDs{})
	r := RecordFactory{}.New("r1", "m1")
	r.Add(laps.New(pace(t, "3:20"), now), laps.New(pace(t, "3:25"), now), laps.New(pace(t, "3:30"), now))

	if err := r.Edit("lap2", pace(t, "3:15")); err != nil {
		t.Fatal(err)
	}
	if err := r.Delete("lap1"); err != nil {
		t.Fatal(err)
	}
	if got := strings(r.Times()); len(got) != 2 || got[0] != "3:15" || got[1] != "3:30" {
		t.Errorf("Times() = %v, want [3:15 3:30]", got)
	}
	if err := r.Edit("missing", pace(t, "3:00")); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Edit(missing) = %v, want ErrNotFound", err)
	}
	if err := r.Delete("missing"); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Delete(missing) = %v, want ErrNotFound", err)
	}
}

func TestPolicies(t *testing.T) {
	if err := (SubmissionPolicy{}).CanChange(false); !errors.Is(err, ErrNotRunner) {
		t.Errorf("CanChange(false) = %v, want ErrNotRunner", err)
	}
	if err := (SubmissionPolicy{}).CanChange(true); err != nil {
		t.Errorf("CanChange(true) = %v", err)
	}

	v := VisibilityPolicy{}
	if !v.CanSee("m1", false, "m1") || v.CanSee("m2", false, "m1") || !v.CanSee("", true, "m1") || v.CanSee("", false, "") {
		t.Error("CanSee() is wrong")
	}
}
