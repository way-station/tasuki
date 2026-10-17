package shared

import "time"

const dateLayout = "2006-01-02"

// Date is a calendar date without a time or a time zone, such as a race date.
type Date struct {
	t time.Time // midnight UTC of the date
}

// ParseDate parses a date written as YYYY-MM-DD.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, NewValidationError("date", "must be YYYY-MM-DD")
	}
	return Date{t: t}, nil
}

// DateIn returns the calendar date of t in loc.
func DateIn(t time.Time, loc *time.Location) Date {
	y, m, d := t.In(loc).Date()
	return Date{t: time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}

// String returns the date as YYYY-MM-DD.
func (d Date) String() string {
	return d.t.Format(dateLayout)
}

// Before reports whether d is before other.
func (d Date) Before(other Date) bool {
	return d.t.Before(other.t)
}

// IsZero reports whether d is the zero Date.
func (d Date) IsZero() bool {
	return d.t.IsZero()
}
