package anomaly

import (
	"testing"
	"time"
)

func baseline(yesterdayAI uint64) Inputs {
	per := map[string]map[string]uint64{}
	dates := []string{}
	today := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	for i := 14; i >= 1; i-- {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, d)
		per[d] = map[string]uint64{
			"ai":     1 * 60 * 60 * 1000,
			"manual": 2 * 60 * 60 * 1000,
		}
	}
	yesterday := today.AddDate(0, 0, -0).Format("2006-01-02")
	dates = append(dates, yesterday)
	per[yesterday] = map[string]uint64{
		"ai":     yesterdayAI,
		"manual": 2 * 60 * 60 * 1000,
	}
	return Inputs{PerCategory: per, Dates: dates}
}

func TestDetect_NoAnomaly(t *testing.T) {
	out := Detect(baseline(1 * 60 * 60 * 1000))
	if len(out) != 0 {
		t.Errorf("expected no anomalies, got %+v", out)
	}
}

func TestDetect_AIHigh(t *testing.T) {

	in := baseline(0)

	for i, d := range in.Dates[:14] {
		in.PerCategory[d]["ai"] = uint64(60*60*1000) + uint64(i*60*1000)
	}
	in.PerCategory[in.Dates[14]]["ai"] = 5 * 60 * 60 * 1000
	out := Detect(in)
	found := false
	for _, a := range out {
		if a.Kind == KindAIHigh {
			found = true
			if a.ZScore < 2.0 {
				t.Errorf("z=%v, expected ≥2.0", a.ZScore)
			}
		}
	}
	if !found {
		t.Errorf("expected ai_high, got %+v", out)
	}
}

func TestDetect_AILow(t *testing.T) {
	in := baseline(0)
	for i, d := range in.Dates[:14] {
		in.PerCategory[d]["ai"] = uint64(60*60*1000) + uint64(i*60*1000)
	}
	in.PerCategory[in.Dates[14]]["ai"] = 0
	out := Detect(in)
	found := false
	for _, a := range out {
		if a.Kind == KindAILow {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ai_low, got %+v", out)
	}
}

func TestDetect_ActivityLow(t *testing.T) {
	in := baseline(60 * 60 * 1000)

	for i, d := range in.Dates[:14] {
		in.PerCategory[d]["ai"] = uint64(60*60*1000) + uint64(i*60*1000)
		in.PerCategory[d]["manual"] = uint64(2*60*60*1000) + uint64(i*60*1000)
	}

	in.PerCategory[in.Dates[14]]["ai"] = 0
	in.PerCategory[in.Dates[14]]["manual"] = 0
	out := Detect(in)
	hasActivityLow := false
	for _, a := range out {
		if a.Kind == KindActivityLow {
			hasActivityLow = true
		}
	}
	if !hasActivityLow {
		t.Errorf("expected activity_low, got %+v", out)
	}
}

func TestDetect_NoiseFloor(t *testing.T) {

	per := map[string]map[string]uint64{}
	dates := []string{}
	today := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	for i := 14; i >= 0; i-- {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, d)
		per[d] = map[string]uint64{"ai": 60 * 1000}
	}
	per[dates[14]]["ai"] = 30 * 60 * 1000
	out := Detect(Inputs{PerCategory: per, Dates: dates})
	for _, a := range out {
		if a.Kind == KindAIHigh {
			t.Errorf("noise floor breached: %+v", a)
		}
	}
}

func TestDetect_TooFewDays(t *testing.T) {
	per := map[string]map[string]uint64{
		"2026-05-09": {"ai": 1 * 60 * 60 * 1000},
		"2026-05-10": {"ai": 100 * 60 * 60 * 1000},
	}
	out := Detect(Inputs{PerCategory: per, Dates: []string{"2026-05-09", "2026-05-10"}})
	if len(out) != 0 {
		t.Errorf("expected no anomalies (too few days), got %+v", out)
	}
}

func TestMeanStd(t *testing.T) {
	mean, std := meanStd([]float64{1, 2, 3, 4, 5})
	if mean != 3 {
		t.Errorf("mean=%v, want 3", mean)
	}
	if std < 1.5 || std > 1.6 {
		t.Errorf("std=%v, want ~1.581", std)
	}
}

func TestZScore_ZeroStd(t *testing.T) {
	if z := zScore(100, 50, 0); z != 0 {
		t.Errorf("zScore with std=0 should be 0, got %v", z)
	}
}

func TestMakeInputs_FillsMissingDates(t *testing.T) {
	today := time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)
	pts := []Trend{
		{Date: "2026-05-09", Category: "ai", MS: 1000},
	}
	in := MakeInputs(pts, today)
	if len(in.Dates) != 15 {
		t.Errorf("dates len=%d, want 15 (14 baseline + yesterday)", len(in.Dates))
	}

	for _, d := range in.Dates {
		if in.PerCategory[d] == nil {
			t.Errorf("missing per-cat map for %s", d)
		}
	}
}

func TestRound2(t *testing.T) {
	if round2(1.234567) != 1.23 {
		t.Error("round2")
	}
	if round2(-2.567) != -2.57 {
		t.Errorf("got %v", round2(-2.567))
	}
}
