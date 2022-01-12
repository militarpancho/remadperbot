package miscelanea

import (
	"time"
)

const newLayout = "15:04"

func CheckOpenGreenPoints() bool {
	now := time.Now()
	today3pm, _ := time.Parse(newLayout, "14:00")
	today8pm, _ := time.Parse(newLayout, "19:00")
	return inTimeSpan(today3pm, today8pm, now)
}

func SecondsICanSleep() int {
	now := time.Now().Format(newLayout)
	clock, _ := time.Parse(newLayout, now)
	today3pm, _ := time.Parse(newLayout, "01:00")
	return int(today3pm.Sub(clock).Seconds())
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
