package util

import "time"

func ClearHourMinuteSecond(t time.Time) time.Time {
	return time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0,
		0,
		0,
		0,
		t.Location(),
	)
}

/// AreDatesIntersecting check if the date-range A intersects with date-range B.
/// Since this is for DATES and not DATETIMES, the arguments are normalized
//inside / the function first (hour-minute-second is culled).
func AreDatesIntersecting(startA, endA, startB, endB time.Time) bool {
	startA = ClearHourMinuteSecond(startA)
	endA = ClearHourMinuteSecond(endA)
	startB = ClearHourMinuteSecond(startB)
	endB = ClearHourMinuteSecond(endB)

	return startA.Before(endB) || startA.Equal(endB) &&
		startB.Before(endA) || startB.Equal(endA)
}
