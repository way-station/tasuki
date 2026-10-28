package shared

import "time"

const dateLayout = "2006-01-02"

// Date is a calendar date without a time or a time zone, such as a race date.
type Date struct {
	midnightUTC time.Time // midnight UTC of the date
}

// ParseDate parses a date written as YYYY-MM-DD.
func ParseDate(text string) (Date, error) {
	parsed, err := time.Parse(dateLayout, text)
	if err != nil {
		return Date{}, NewValidationError("date", "must be YYYY-MM-DD")
	}
	return Date{midnightUTC: parsed}, nil
}

// DateIn returns the calendar date of moment in location.
func DateIn(moment time.Time, location *time.Location) Date {
	year, month, day := moment.In(location).Date()
	return Date{midnightUTC: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// String returns the date as YYYY-MM-DD.
func (date Date) String() string {
	return date.midnightUTC.Format(dateLayout)
}

// Before reports whether date is before other.
func (date Date) Before(other Date) bool {
	return date.midnightUTC.Before(other.midnightUTC)
}

// IsZero reports whether date is the zero Date.
func (date Date) IsZero() bool {
	return date.midnightUTC.IsZero()
}
