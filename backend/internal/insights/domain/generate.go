package domain

import (
	"sort"
	"time"
)

func Generate(in Inputs) []Insight {
	out := []Insight{}
	if x := aiTrendInsight(in); x != nil {
		out = append(out, *x)
	}
	if x := aiRatioInsight(in); x != nil {
		out = append(out, *x)
	}
	if x := topLangInsight(in); x != nil {
		out = append(out, *x)
	}
	if x := productiveDayInsight(in); x != nil {
		out = append(out, *x)
	}
	if x := totalActivityInsight(in); x != nil {
		out = append(out, *x)
	}
	return out
}

func aiTrendInsight(in Inputs) *Insight {
	const minMS = 5 * 60 * 1000
	last := float64(in.Last7d["ai"])
	prev := float64(in.Prev7d["ai"])
	if last < minMS && prev < minMS {
		return nil
	}
	if prev < minMS {
		return &Insight{Key: "ai_trend_started", Vars: map[string]any{"hours": round1(last / 3600000)}}
	}
	pct := int((last - prev) * 100 / prev)
	if pct >= 10 {
		return &Insight{Key: "ai_trend_up", Vars: map[string]any{"pct": pct}}
	}
	if pct <= -10 {
		return &Insight{Key: "ai_trend_down", Vars: map[string]any{"pct": -pct}}
	}
	return &Insight{Key: "ai_trend_flat", Vars: nil}
}

func aiRatioInsight(in Inputs) *Insight {
	ai := in.Last7d["ai"]
	coding := ai + in.Last7d["manual"] + in.Last7d["refactor"]
	const minMS = 30 * 60 * 1000
	if coding < minMS {
		return nil
	}
	pct := int(ai * 100 / coding)
	return &Insight{Key: "ai_ratio", Vars: map[string]any{"pct": pct}}
}

func topLangInsight(in Inputs) *Insight {
	totals := make(map[string]uint64)
	var grand uint64
	for _, c := range in.Languages {
		totals[c.Lang] += c.MS
		grand += c.MS
	}
	if grand == 0 {
		return nil
	}
	type kv struct {
		lang string
		ms   uint64
	}
	all := make([]kv, 0, len(totals))
	for l, ms := range totals {
		if l == "" {
			continue
		}
		all = append(all, kv{l, ms})
	}
	if len(all) == 0 {
		return nil
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ms > all[j].ms })
	top := all[0]
	pct := int(top.ms * 100 / grand)
	return &Insight{Key: "top_lang", Vars: map[string]any{"lang": top.lang, "pct": pct}}
}

func productiveDayInsight(in Inputs) *Insight {
	if len(in.Trend) == 0 {
		return nil
	}
	byDate := make(map[string]uint64)
	for _, p := range in.Trend {
		if p.Category == "ai" || p.Category == "manual" || p.Category == "refactor" {
			byDate[p.Date] += p.MS
		}
	}
	if len(byDate) == 0 {
		return nil
	}
	var bestDate string
	var bestMS uint64
	for d, ms := range byDate {
		if ms > bestMS {
			bestMS = ms
			bestDate = d
		}
	}
	if bestMS < 30*60*1000 {
		return nil
	}
	t, err := time.Parse("2006-01-02", bestDate)
	if err != nil {
		return nil
	}
	return &Insight{
		Key:  "productive_day",
		Vars: map[string]any{"dow": int(t.Weekday()), "hours": round1(float64(bestMS) / 3600000)},
	}
}

func totalActivityInsight(in Inputs) *Insight {
	total := in.Last7d["ai"] + in.Last7d["manual"] + in.Last7d["refactor"]
	if total == 0 {
		return nil
	}
	hours := round1(float64(total) / 3600000)
	if hours < 0.5 {
		return nil
	}
	return &Insight{Key: "total_activity", Vars: map[string]any{"hours": hours}}
}

func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
