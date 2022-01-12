package miscelanea

import (
	"time"
)

const newLayout = "15:04"

func CheckOpenGreenPoints() bool {
	now := time.Now().Format(newLayout)
	clock, _ := time.Parse(newLayout, now)
	today8pm, _ := time.Parse(newLayout, "07:00")
	today20pm, _ := time.Parse(newLayout, "19:00")
	return inTimeSpan(today8pm, today20pm, clock)
}

func SecondsICanSleep() int {
	now := time.Now().Format(newLayout)
	clock, _ := time.Parse(newLayout, now)
	today8pm, _ := time.Parse(newLayout, "07:00")
	return int(today8pm.Sub(clock).Seconds())
}

func inTimeSpan(start, end, check time.Time) bool {
	if start.Before(end) {
		return !check.Before(start) && !check.After(end)
	}
	if start.Equal(end) {
		return check.Equal(start)
	}
	return !start.After(check) || !end.Before(check)
}
