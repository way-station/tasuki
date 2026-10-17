package shared

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	d, err := ParseDate("2026-11-23")
	if err != nil {
		t.Fatal(err)
	}
	if d.String() != "2026-11-23" {
		t.Errorf("String() = %q", d.String())
	}
	if _, err := ParseDate("2026/11/23"); err == nil {
		t.Error("ParseDate(2026/11/23): want an error")
	}
}

func TestDateIn(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	// 2026-11-22 20:00 UTC is already 2026-11-23 in Japan.
	got := DateIn(time.Date(2026, 11, 22, 20, 0, 0, 0, time.UTC), jst)
	if got.String() != "2026-11-23" {
		t.Errorf("DateIn() = %s, want 2026-11-23", got)
	}
	before, _ := ParseDate("2026-11-22")
	if !before.Before(got) || got.Before(before) {
		t.Error("Before() is wrong")
	}
}
