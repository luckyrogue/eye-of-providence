// Package insights — превращает сырые агрегаты ClickHouse в
// пользовательские narrative-карточки (i18n key + template vars).
//
// Frontend получает {key, vars} и сам резолвит локализованную строку через
// i18next interpolation. Это разделяет копирайтинг (фронт) и логику отбора
// (бэк), плюс автоматически работает в RU/EN/KK/ES.
package insights

import (
	"sort"
	"time"

	"github.com/eye-of-providence/backend/internal/store"
)

// Insight — единичная карточка для UI. Frontend i18n key resolve через
// `t(insights:${key}, vars)`. Vars — JSON-serializable (number/string/bool).
type Insight struct {
	Key  string         `json:"key"`
	Vars map[string]any `json:"vars,omitempty"`
}

// Inputs — данные для генерации. Все aggregates over одного user'а.
type Inputs struct {
	// Last7d / Prev7d — категория → ms за last 7 дней / 7-14 дней назад.
	Last7d, Prev7d map[string]uint64
	// Languages — last 30d, lang × category.
	Languages []store.LangCell
	// Trend — last 7d по дням × категория.
	Trend []store.TrendPoint
}

// Generate — детерминистичный отбор insights в фиксированном порядке.
// Каждый sub-generator может вернуть 0 или 1 insight.
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

// aiTrendInsight — изменение AI-времени неделя-к-неделе.
//
//	ai_trend_up   — этой неделей AI на N% больше
//	ai_trend_down — на N% меньше
//	ai_trend_flat — без значимого изменения (порог ±10%)
//
// Без insight'а если в обеих неделях AI <5 минут (статистический шум).
func aiTrendInsight(in Inputs) *Insight {
	const minMS = 5 * 60 * 1000 // 5 минут
	last := float64(in.Last7d["ai"])
	prev := float64(in.Prev7d["ai"])
	if last < minMS && prev < minMS {
		return nil
	}
	if prev < minMS {
		// ai только начал использоваться — отдельный insight
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

// aiRatioInsight — какая доля кода написана с AI за последнюю неделю.
// "Coding time" = ai+manual+refactor (idle/reading/other не считаем).
func aiRatioInsight(in Inputs) *Insight {
	ai := in.Last7d["ai"]
	coding := ai + in.Last7d["manual"] + in.Last7d["refactor"]
	const minMS = 30 * 60 * 1000 // 30 минут
	if coding < minMS {
		return nil
	}
	pct := int(ai * 100 / coding)
	return &Insight{Key: "ai_ratio", Vars: map[string]any{"pct": pct}}
}

// topLangInsight — самый используемый язык за last 30d.
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

// productiveDayInsight — день недели с максимальным coding time
// за последние 7 дней.
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
	if bestMS < 30*60*1000 { // <30 мин — не показываем
		return nil
	}
	t, err := time.Parse("2006-01-02", bestDate)
	if err != nil {
		return nil
	}
	return &Insight{
		Key: "productive_day",
		Vars: map[string]any{
			// dow 0..6, frontend сам резолвит локализованное название
			"dow":   int(t.Weekday()),
			"hours": round1(float64(bestMS) / 3600000),
		},
	}
}

// totalActivityInsight — общий coding-time за последние 7 дней.
// Базовая метрика, всегда показывается если есть активность.
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

// round1 — округление до одного знака после запятой.
func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
