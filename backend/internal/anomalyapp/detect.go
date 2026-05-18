package anomalyapp

import (
	"time"

	"github.com/eye-of-providence/backend/internal/anomaly"
)

func DetectTrends(points []TrendPoint, today time.Time) []anomaly.Anomaly {
	trends := make([]anomaly.Trend, 0, len(points))
	for _, p := range points {
		trends = append(trends, anomaly.Trend{Date: p.Date, Category: p.Category, MS: p.MS})
	}
	return anomaly.Detect(anomaly.MakeInputs(trends, today))
}
