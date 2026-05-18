package domain

import "time"

// DaysWindow — bounded lookback for aggregations.
type DaysWindow struct {
	Days  int
	Since time.Time
}

// TrendQuery — daily trend parameters.
type TrendQuery struct {
	DaysWindow
	TZ string
}
