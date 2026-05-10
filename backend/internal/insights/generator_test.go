package insights

import (
	"testing"

	"github.com/eye-of-providence/backend/internal/store"
)

func TestAITrend_Up(t *testing.T) {
	in := Inputs{
		Last7d: map[string]uint64{"ai": 12 * 60 * 60 * 1000},
		Prev7d: map[string]uint64{"ai": 10 * 60 * 60 * 1000},
	}
	got := aiTrendInsight(in)
	if got == nil || got.Key != "ai_trend_up" {
		t.Fatalf("got %+v, want ai_trend_up", got)
	}
	if pct, _ := got.Vars["pct"].(int); pct != 20 {
		t.Errorf("pct=%v, want 20", got.Vars["pct"])
	}
}

func TestAITrend_Down(t *testing.T) {
	in := Inputs{
		Last7d: map[string]uint64{"ai": 5 * 60 * 60 * 1000},
		Prev7d: map[string]uint64{"ai": 10 * 60 * 60 * 1000},
	}
	got := aiTrendInsight(in)
	if got == nil || got.Key != "ai_trend_down" {
		t.Fatalf("got %+v, want ai_trend_down", got)
	}
	if pct, _ := got.Vars["pct"].(int); pct != 50 {
		t.Errorf("pct=%v, want 50", got.Vars["pct"])
	}
}

func TestAITrend_Flat(t *testing.T) {
	// 5% change → flat
	in := Inputs{
		Last7d: map[string]uint64{"ai": 105 * 60 * 1000},
		Prev7d: map[string]uint64{"ai": 100 * 60 * 1000},
	}
	got := aiTrendInsight(in)
	if got == nil || got.Key != "ai_trend_flat" {
		t.Errorf("got %+v, want ai_trend_flat", got)
	}
}

func TestAITrend_Started(t *testing.T) {
	in := Inputs{
		Last7d: map[string]uint64{"ai": 60 * 60 * 1000},
		Prev7d: map[string]uint64{"ai": 0},
	}
	got := aiTrendInsight(in)
	if got == nil || got.Key != "ai_trend_started" {
		t.Fatalf("got %+v, want ai_trend_started", got)
	}
	if h := got.Vars["hours"].(float64); h != 1.0 {
		t.Errorf("hours=%v, want 1.0", h)
	}
}

func TestAITrend_NoiseSuppressed(t *testing.T) {
	in := Inputs{
		Last7d: map[string]uint64{"ai": 60 * 1000},
		Prev7d: map[string]uint64{"ai": 60 * 1000},
	}
	if got := aiTrendInsight(in); got != nil {
		t.Errorf("got %+v, want nil (below noise floor)", got)
	}
}

func TestAIRatio(t *testing.T) {
	in := Inputs{
		Last7d: map[string]uint64{
			"ai":       30 * 60 * 1000,
			"manual":   60 * 60 * 1000,
			"refactor": 10 * 60 * 1000,
		},
	}
	got := aiRatioInsight(in)
	if got == nil || got.Key != "ai_ratio" {
		t.Fatalf("got %+v", got)
	}
	if pct, _ := got.Vars["pct"].(int); pct != 30 {
		t.Errorf("pct=%v, want 30", pct)
	}
}

func TestAIRatio_NoCoding(t *testing.T) {
	in := Inputs{Last7d: map[string]uint64{"reading": 10 * 60 * 1000}}
	if got := aiRatioInsight(in); got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestTopLang(t *testing.T) {
	in := Inputs{
		Languages: []store.LangCell{
			{Lang: "go", MS: 4 * 60 * 60 * 1000, Category: "manual"},
			{Lang: "python", MS: 6 * 60 * 60 * 1000, Category: "manual"},
			{Lang: "python", MS: 2 * 60 * 60 * 1000, Category: "ai"},
		},
	}
	got := topLangInsight(in)
	if got == nil || got.Key != "top_lang" {
		t.Fatalf("got %+v", got)
	}
	if l, _ := got.Vars["lang"].(string); l != "python" {
		t.Errorf("lang=%v, want python", l)
	}
	// python = 8h из 12h total = 66%
	if pct, _ := got.Vars["pct"].(int); pct < 60 || pct > 70 {
		t.Errorf("pct=%v, want 60-70", pct)
	}
}

func TestTopLang_SkipsEmpty(t *testing.T) {
	in := Inputs{
		Languages: []store.LangCell{
			{Lang: "", MS: 100 * 60 * 60 * 1000, Category: "manual"},
			{Lang: "rust", MS: 1 * 60 * 60 * 1000, Category: "manual"},
		},
	}
	got := topLangInsight(in)
	if got == nil || got.Vars["lang"] != "rust" {
		t.Errorf("got %+v, want rust (empty lang skipped)", got)
	}
}

func TestProductiveDay(t *testing.T) {
	in := Inputs{
		Trend: []store.TrendPoint{
			{Date: "2026-05-04", Category: "manual", MS: 1 * 60 * 60 * 1000}, // Monday
			{Date: "2026-05-08", Category: "manual", MS: 5 * 60 * 60 * 1000}, // Friday
			{Date: "2026-05-08", Category: "ai", MS: 1 * 60 * 60 * 1000},     // Friday +1h
		},
	}
	got := productiveDayInsight(in)
	if got == nil || got.Key != "productive_day" {
		t.Fatalf("got %+v", got)
	}
	// Friday = 5 (Mon=1, Tue=2, ..., Fri=5, Sat=6, Sun=0)
	if dow, _ := got.Vars["dow"].(int); dow != 5 {
		t.Errorf("dow=%v, want 5 (Friday)", dow)
	}
}

func TestProductiveDay_BelowThreshold(t *testing.T) {
	in := Inputs{
		Trend: []store.TrendPoint{
			{Date: "2026-05-08", Category: "manual", MS: 5 * 60 * 1000},
		},
	}
	if got := productiveDayInsight(in); got != nil {
		t.Errorf("got %+v, want nil (below 30min threshold)", got)
	}
}

func TestTotalActivity(t *testing.T) {
	in := Inputs{
		Last7d: map[string]uint64{
			"ai":     2 * 60 * 60 * 1000,
			"manual": 3 * 60 * 60 * 1000,
		},
	}
	got := totalActivityInsight(in)
	if got == nil || got.Key != "total_activity" {
		t.Fatalf("got %+v", got)
	}
	if h, _ := got.Vars["hours"].(float64); h != 5.0 {
		t.Errorf("hours=%v, want 5.0", h)
	}
}

func TestGenerate_DeterministicOrder(t *testing.T) {
	in := Inputs{
		Last7d: map[string]uint64{
			"ai":     2 * 60 * 60 * 1000,
			"manual": 3 * 60 * 60 * 1000,
		},
		Prev7d: map[string]uint64{"ai": 1 * 60 * 60 * 1000},
		Languages: []store.LangCell{
			{Lang: "go", MS: 5 * 60 * 60 * 1000, Category: "manual"},
		},
		Trend: []store.TrendPoint{
			{Date: "2026-05-08", Category: "manual", MS: 4 * 60 * 60 * 1000},
		},
	}
	out := Generate(in)
	if len(out) < 4 {
		t.Fatalf("got %d insights, want >=4", len(out))
	}
	// Order: ai_trend → ai_ratio → top_lang → productive_day → total_activity
	wantOrder := []string{"ai_trend_up", "ai_ratio", "top_lang", "productive_day", "total_activity"}
	for i, k := range wantOrder {
		if out[i].Key != k {
			t.Errorf("position %d: got %s, want %s", i, out[i].Key, k)
		}
	}
}

func TestGenerate_NoData_NoInsights(t *testing.T) {
	out := Generate(Inputs{})
	if len(out) != 0 {
		t.Errorf("got %d insights for empty input, want 0", len(out))
	}
}
