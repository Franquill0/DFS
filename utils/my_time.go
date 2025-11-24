package my_time

import (
	"fmt"
	"time"
)

func Now() time.Time {
	return time.Now()
}
func Since(start time.Time) time.Duration {
	return time.Since(start)
}
func GetFormattedTime(timeStart time.Time) string {
	timeDuration := time.Since(timeStart)
	minutes := int(timeDuration.Minutes())
	seconds := int(timeDuration.Seconds()) % 60
	return fmt.Sprintf("%dmin %ds", minutes, seconds)
}
