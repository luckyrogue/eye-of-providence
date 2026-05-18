package anomalyapp

import "github.com/eye-of-providence/backend/internal/anomaly"

func PushPayload(a anomaly.Anomaly) any {
	titles := map[anomaly.Kind]string{
		anomaly.KindAIHigh:       "AI usage spike",
		anomaly.KindAILow:        "AI usage dip",
		anomaly.KindManualHigh:   "Manual coding spike",
		anomaly.KindManualLow:    "Manual coding dip",
		anomaly.KindRefactorHigh: "Refactoring day",
		anomaly.KindActivityHigh: "Productivity spike",
		anomaly.KindActivityLow:  "Activity dropped",
	}
	title, ok := titles[a.Kind]
	if !ok {
		title = "Coding anomaly"
	}
	yHrs := float64(a.YesterdayMS) / 3600000
	bHrs := float64(a.BaselineMS) / 3600000
	return map[string]any{
		"title": title,
		"body":  formatAnomalyBody(yHrs, bHrs, a.Category),
		"url":   "/dashboard",
		"tag":   "anomaly." + string(a.Kind),
	}
}

func formatAnomalyBody(yHrs, bHrs float64, category string) string {
	return formatHours(yHrs) + " vs " + formatHours(bHrs) + " baseline (" + category + ")"
}

func formatHours(h float64) string {
	if h < 1 {
		min := int(h*60 + 0.5)
		return itoa(min) + "min"
	}
	tenths := int(h*10 + 0.5)
	whole := tenths / 10
	frac := tenths % 10
	if frac == 0 {
		return itoa(whole) + "h"
	}
	return itoa(whole) + "." + itoa(frac) + "h"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
