package util

import (
	"log"
	"time"
)

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

// AreDatesIntersecting checks if date-range A intersects with date-range B.
// Dates are normalized to remove time components (hour, minute, second).
func AreDatesIntersecting(startA, endA, startB, endB time.Time) bool {
	startA = ClearHourMinuteSecond(startA)
	endA = ClearHourMinuteSecond(endA)
	startB = ClearHourMinuteSecond(startB)
	endB = ClearHourMinuteSecond(endB)

	log.Printf("Comparing: %s - %s with %s - %s", startA.Format("2006-01-02"), endA.Format("2006-01-02"), startB.Format("2006-01-02"), endB.Format("2006-01-02"))

	// Ensure valid ranges
	if endA.Before(startA) || endB.Before(startB) {
		log.Printf("Invalid date range detected")
		return false
	}

	// Check for intersection: A starts before B ends AND B starts before A ends
	if (startA.Before(endB) || startA.Equal(endB)) &&
		(startB.Before(endA) || startB.Equal(endA)) {
		return true
	}

	return false
}
