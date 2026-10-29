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
func ParsePace(text string) (Pace, error) {
	match := paceFormat.FindStringSubmatch(text)
	if match == nil {
		return Pace{}, NewValidationError("time", "must be MM:SS")
	}
	minutes, _ := strconv.Atoi(match[1])
	secs, _ := strconv.Atoi(match[2])
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
func (pace Pace) Seconds() int {
	return pace.seconds
}

// String returns the pace as MM:SS (for example "3:28").
func (pace Pace) String() string {
	return fmt.Sprintf("%d:%02d", pace.seconds/60, pace.seconds%60)
}

// TotalTime is a time written as H:MM:SS, such as a target or a predicted total time.
type TotalTime struct {
	seconds int
}

// ParseTotalTime parses a time written as H:MM:SS (for example "2:45:30").
func ParseTotalTime(text string) (TotalTime, error) {
	match := totalTimeFormat.FindStringSubmatch(text)
	if match == nil {
		return TotalTime{}, NewValidationError("total_time", "must be H:MM:SS")
	}
	hours, err := strconv.Atoi(match[1])
	if err != nil {
		return TotalTime{}, NewValidationError("total_time", "must be H:MM:SS")
	}
	minutes, _ := strconv.Atoi(match[2])
	seconds, _ := strconv.Atoi(match[3])
	return TotalTimeFromSeconds(hours*3600 + minutes*60 + seconds)
}

// TotalTimeFromSeconds returns the total time for a whole number of seconds.
func TotalTimeFromSeconds(seconds int) (TotalTime, error) {
	if seconds <= 0 {
		return TotalTime{}, NewValidationError("total_time", "must be greater than 0:00:00")
	}
	return TotalTime{seconds: seconds}, nil
}

// Seconds returns the total time in seconds.
func (totalTime TotalTime) Seconds() int {
	return totalTime.seconds
}

// String returns the time as H:MM:SS (for example "2:45:30").
func (totalTime TotalTime) String() string {
	return fmt.Sprintf("%d:%02d:%02d", totalTime.seconds/3600, totalTime.seconds/60%60, totalTime.seconds%60)
}
