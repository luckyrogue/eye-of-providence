package anomaly

import (
	"math"
	"time"
)

type Kind string

const (
	KindAIHigh       Kind = "ai_high"
	KindAILow        Kind = "ai_low"
	KindManualHigh   Kind = "manual_high"
	KindManualLow    Kind = "manual_low"
	KindRefactorHigh Kind = "refactor_high"
	KindActivityHigh Kind = "activity_high"
	KindActivityLow  Kind = "activity_low"
)

type Anomaly struct {
	Kind        Kind    `json:"kind"`
	Category    string  `json:"category"`
	ZScore      float64 `json:"z_score"`
	YesterdayMS uint64  `json:"yesterday_ms"`
	BaselineMS  uint64  `json:"baseline_ms"`
	Date        string  `json:"date"`
}

type Inputs struct {
	PerCategory map[string]map[string]uint64

	Dates []string
}

const (
	minBaselineDays = 7

	zThreshold = 2.0

	minBaselineMS = 5 * 60 * 1000
)

func Detect(in Inputs) []Anomaly {
	if len(in.Dates) < minBaselineDays+1 {
		return nil
	}
	yesterday := in.Dates[len(in.Dates)-1]
	baselineDates := in.Dates[:len(in.Dates)-1]
	if len(baselineDates) > 14 {
		baselineDates = baselineDates[len(baselineDates)-14:]
	}

	out := []Anomaly{}
	for _, cat := range []string{"ai", "manual", "refactor"} {
		if a := detectCategory(in.PerCategory, baselineDates, yesterday, cat); a != nil {
			out = append(out, *a)
		}
	}
	if a := detectTotal(in.PerCategory, baselineDates, yesterday); a != nil {
		out = append(out, *a)
	}
	return out
}

func detectCategory(per map[string]map[string]uint64, baseline []string, yesterday, cat string) *Anomaly {
	values := make([]float64, len(baseline))
	for i, d := range baseline {
		values[i] = float64(per[d][cat])
	}
	mean, std := meanStd(values)
	if mean < float64(minBaselineMS) {
		return nil
	}
	yMS := per[yesterday][cat]
	z := zScore(float64(yMS), mean, std)
	if math.Abs(z) < zThreshold {
		return nil
	}
	kind := categoryKind(cat, z > 0)
	if kind == "" {
		return nil
	}
	return &Anomaly{
		Kind:        kind,
		Category:    cat,
		ZScore:      round2(z),
		YesterdayMS: yMS,
		BaselineMS:  uint64(mean),
		Date:        yesterday,
	}
}

func detectTotal(per map[string]map[string]uint64, baseline []string, yesterday string) *Anomaly {
	values := make([]float64, len(baseline))
	for i, d := range baseline {
		values[i] = float64(totalCoding(per[d]))
	}
	mean, std := meanStd(values)
	if mean < float64(minBaselineMS) {
		return nil
	}
	yMS := totalCoding(per[yesterday])
	z := zScore(float64(yMS), mean, std)
	if math.Abs(z) < zThreshold {
		return nil
	}
	kind := KindActivityHigh
	if z < 0 {
		kind = KindActivityLow
	}
	return &Anomaly{
		Kind:        kind,
		Category:    "total",
		ZScore:      round2(z),
		YesterdayMS: yMS,
		BaselineMS:  uint64(mean),
		Date:        yesterday,
	}
}

func totalCoding(day map[string]uint64) uint64 {
	return day["ai"] + day["manual"] + day["refactor"]
}

func categoryKind(cat string, high bool) Kind {
	switch cat {
	case "ai":
		if high {
			return KindAIHigh
		}
		return KindAILow
	case "manual":
		if high {
			return KindManualHigh
		}
		return KindManualLow
	case "refactor":
		if high {
			return KindRefactorHigh
		}

		return ""
	}
	return ""
}

func meanStd(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	if len(xs) < 2 {
		return mean, 0
	}
	var ss float64
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	return mean, math.Sqrt(ss / float64(len(xs)-1))
}

func zScore(x, mean, std float64) float64 {
	if std == 0 {
		return 0
	}
	return (x - mean) / std
}

type Trend struct {
	Date     string
	Category string
	MS       uint64
}

func MakeInputs(points []Trend, today time.Time) Inputs {
	per := map[string]map[string]uint64{}
	for _, p := range points {
		if per[p.Date] == nil {
			per[p.Date] = map[string]uint64{}
		}
		per[p.Date][p.Category] += p.MS
	}

	yesterday := today.UTC().AddDate(0, 0, -1)
	dates := []string{}
	for i := 14; i >= 0; i-- {
		d := yesterday.AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, d)
		if per[d] == nil {
			per[d] = map[string]uint64{}
		}
	}
	return Inputs{PerCategory: per, Dates: dates}
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}
