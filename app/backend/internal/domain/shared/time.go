package shared

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	paceFormat      = regexp.MustCompile(`^([0-9]{1,2}):([0-5][0-9])$`)
	totalTimeFormat = regexp.MustCompile(`^([0-9]+):([0-5][0-9]):([0-5][0-9])$`)
)

// Pace is a time per km written as MM:SS, such as a lap time or an average time.
// It is held as whole seconds (docs/design/database.md).
type Pace struct {
	seconds int
}

// ParsePace parses a pace written as MM:SS (for example "3:28").
func ParsePace(s string) (Pace, error) {
	m := paceFormat.FindStringSubmatch(s)
	if m == nil {
		return Pace{}, NewValidationError("time", "must be MM:SS")
	}
	minutes, _ := strconv.Atoi(m[1])
	secs, _ := strconv.Atoi(m[2])
	return PaceFromSeconds(minutes*60 + secs)
}

// PaceFromSeconds returns the pace for a whole number of seconds per km.
func PaceFromSeconds(seconds int) (Pace, error) {
	if seconds <= 0 || seconds >= 100*60 {
		return Pace{}, NewValidationError("time", "must be between 0:01 and 99:59")
	}
	return Pace{seconds: seconds}, nil
}

// Seconds returns the pace in seconds per km.
func (p Pace) Seconds() int {
	return p.seconds
}

// String returns the pace as MM:SS (for example "3:28").
func (p Pace) String() string {
	return fmt.Sprintf("%d:%02d", p.seconds/60, p.seconds%60)
}

// TotalTime is a time written as H:MM:SS, such as a target or a predicted total time.
type TotalTime struct {
	seconds int
}

// ParseTotalTime parses a time written as H:MM:SS (for example "2:45:30").
func ParseTotalTime(s string) (TotalTime, error) {
	m := totalTimeFormat.FindStringSubmatch(s)
	if m == nil {
		return TotalTime{}, NewValidationError("total_time", "must be H:MM:SS")
	}
	h, err := strconv.Atoi(m[1])
	if err != nil {
		return TotalTime{}, NewValidationError("total_time", "must be H:MM:SS")
	}
	mm, _ := strconv.Atoi(m[2])
	ss, _ := strconv.Atoi(m[3])
	return TotalTimeFromSeconds(h*3600 + mm*60 + ss)
}

// TotalTimeFromSeconds returns the total time for a whole number of seconds.
func TotalTimeFromSeconds(seconds int) (TotalTime, error) {
	if seconds <= 0 {
		return TotalTime{}, NewValidationError("total_time", "must be greater than 0:00:00")
	}
	return TotalTime{seconds: seconds}, nil
}

// Seconds returns the total time in seconds.
func (t TotalTime) Seconds() int {
	return t.seconds
}

// String returns the time as H:MM:SS (for example "2:45:30").
func (t TotalTime) String() string {
	return fmt.Sprintf("%d:%02d:%02d", t.seconds/3600, t.seconds/60%60, t.seconds%60)
}
