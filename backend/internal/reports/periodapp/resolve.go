package periodapp

import (
	"fmt"
	"time"
)

// Resolve — границы периода для POST /v1/reports/generate (?period=weekly|monthly|daily).
func Resolve(kind string, now time.Time) (from, to time.Time, periodKey string) {
	now = now.UTC()
	switch kind {
	case "monthly":
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0).Add(-time.Second)
		periodKey = fmt.Sprintf("monthly_%04d_%02d", now.Year(), now.Month())
		return from, to, periodKey
	case "daily":
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		to = from.Add(24 * time.Hour).Add(-time.Second)
		periodKey = fmt.Sprintf("daily_%s", from.Format("2006-01-02"))
		return from, to, periodKey
	default: // weekly
		year, week := now.ISOWeek()
		offset := int(now.Weekday())
		if offset == 0 {
			offset = 7
		}
		monday := now.AddDate(0, 0, 1-offset)
		from = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 0, 7).Add(-time.Second)
		periodKey = fmt.Sprintf("weekly_%04d_W%02d", year, week)
		return from, to, periodKey
	}
}
