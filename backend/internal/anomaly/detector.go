// Package anomaly — daily anomaly detection over coding-time series.
//
// Подход: Z-score на последних 14 днях. Если yesterday значительно отклонился
// от baseline mean (|z| ≥ threshold), эмиттим anomaly.detected event которое
// webhooks.Service доставляет подписчикам (типично — Slack).
//
// Используется как retention loop: «Команда сегодня необычно много AI»,
// «вчера был всплеск рефакторинга».
package anomaly

import (
	"math"
	"time"
)

// Kind — тип аномалии. Frontend / Slack-receiver используют для рендеринга
// человеко-читаемого сообщения.
type Kind string

const (
	KindAIHigh       Kind = "ai_high"        // AI-время сильно выше baseline
	KindAILow        Kind = "ai_low"         // AI-время сильно ниже
	KindManualHigh   Kind = "manual_high"    // Manual-время сильно выше
	KindManualLow    Kind = "manual_low"     // Manual-время сильно ниже
	KindRefactorHigh Kind = "refactor_high"  // Refactor сильно выше
	KindActivityHigh Kind = "activity_high"  // Общая активность выше
	KindActivityLow  Kind = "activity_low"   // Активность упала (alert на отсутствие работы)
)

// Anomaly — один detected outlier. Многомерный output: per-category возможны
// несколько аномалий за один день.
type Anomaly struct {
	Kind        Kind    `json:"kind"`
	Category    string  `json:"category"`     // ai | manual | refactor | total
	ZScore      float64 `json:"z_score"`      // округлено до 2 знаков
	YesterdayMS uint64  `json:"yesterday_ms"`
	BaselineMS  uint64  `json:"baseline_ms"` // mean previous 14 days
	Date        string  `json:"date"`         // YYYY-MM-DD UTC of "yesterday"
}

// Inputs — daily series, отсортированный по дате (oldest → newest).
// Series должна содержать минимум 8 точек: 7+ для baseline и 1 yesterday.
type Inputs struct {
	// PerCategory[date][category] = duration_ms за день. Для отсутствующих
	// (date, cat) — 0 (не nil).
	PerCategory map[string]map[string]uint64
	// Dates в порядке возрастания. PerCategory[date] должен быть для каждого.
	Dates []string
}

const (
	// minBaselineDays — нужно ≥7 дней истории чтобы baseline был стабильным.
	minBaselineDays = 7
	// zThreshold — порог срабатывания. 2.0 = ~95% confidence (нормальное
	// распределение). 2.5 для меньшего noise; 1.5 для более чувствительных
	// alerts.
	zThreshold = 2.0
	// minBaselineMS — если baseline mean < 5 минут, статистически шумно;
	// игнорируем (z-score нестабильный на низких данных).
	minBaselineMS = 5 * 60 * 1000
)

// Detect — возвращает все аномалии в "yesterday" (последний элемент Dates)
// относительно baseline (предыдущие 7-14 дней).
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
		// refactor low — не интересный signal, игнорим
		return ""
	}
	return ""
}

// meanStd — sample mean + sample standard deviation (Bessel's correction
// для small N — n-1 в знаменателе).
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

// zScore — (x - mean) / std. Если std == 0 (constant baseline), возвращаем 0
// чтобы не triggerить anomaly на zero-variance noise.
func zScore(x, mean, std float64) float64 {
	if std == 0 {
		return 0
	}
	return (x - mean) / std
}

// MakeInputs — convenience: преобразует store.TrendPoint slice в Inputs.
// Включает все категории, заполняет пропущенные (date, cat) нулями.
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
	// Dates: 14 предыдущих + yesterday. yesterday = today - 1.
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
