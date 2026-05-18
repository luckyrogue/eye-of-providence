package anomalyapp_test

import (
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/anomaly"
	"github.com/eye-of-providence/backend/internal/anomalyapp"
)

func TestDetectTrends_MatchesDetector(t *testing.T) {
	today := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	var trends []anomaly.Trend
	for i := 14; i >= 1; i-- {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		trends = append(trends,
			anomaly.Trend{Date: d, Category: "ai", MS: 60 * 60 * 1000},
			anomaly.Trend{Date: d, Category: "manual", MS: 2 * 60 * 60 * 1000},
		)
	}
	yesterday := today.Format("2006-01-02")
	trends = append(trends,
		anomaly.Trend{Date: yesterday, Category: "ai", MS: 5 * 60 * 60 * 1000},
		anomaly.Trend{Date: yesterday, Category: "manual", MS: 2 * 60 * 60 * 1000},
	)
	pts := make([]anomalyapp.TrendPoint, len(trends))
	for i, tr := range trends {
		pts[i] = anomalyapp.TrendPoint{Date: tr.Date, Category: tr.Category, MS: tr.MS}
	}
	want := anomaly.Detect(anomaly.MakeInputs(trends, today))
	got := anomalyapp.DetectTrends(pts, today)
	if len(got) != len(want) {
		t.Fatalf("len got=%d want=%d", len(got), len(want))
	}
}
