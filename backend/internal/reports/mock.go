package reports

import (
	"fmt"
	"strings"
)

// mockReport — детерминистичный отчёт для dev-режима, когда нет API key.
// Используется, чтобы UI можно было разрабатывать без обращений в Gemini.
func mockReport(nc *NumericContext) string {
	var sb strings.Builder
	totalSec := nc.TotalActiveMS / 1000
	aiSec := nc.ByCategoryMS["ai"] / 1000
	manualSec := nc.ByCategoryMS["manual"] / 1000
	ratio := 0
	if totalSec > 0 {
		ratio = int(float64(aiSec) * 100.0 / float64(totalSec))
	}

	fmt.Fprintf(&sb, "# Weekly report — %s\n\n", nc.Period)
	fmt.Fprintf(&sb, "_Mock report (EOP_GEMINI_API_KEY не задан, реальный Gemini не вызывался)._\n\n")
	fmt.Fprintf(&sb, "## TL;DR\nЗа этот период ты потратил %d мин активного времени, из них %d%% с AI.\n\n", totalSec/60, ratio)

	sb.WriteString("## Ключевые цифры\n")
	fmt.Fprintf(&sb, "- Активное время: %d мин\n", totalSec/60)
	fmt.Fprintf(&sb, "- AI: %d мин (%d%%)\n", aiSec/60, ratio)
	fmt.Fprintf(&sb, "- Manual: %d мин\n", manualSec/60)
	fmt.Fprintf(&sb, "- События обработано: %d\n", nc.EventsTotal)
	sb.WriteString("\n")

	if len(nc.AICharsByProvider) > 0 {
		sb.WriteString("## AI по провайдерам (chars)\n")
		for p, c := range nc.AICharsByProvider {
			fmt.Fprintf(&sb, "- %s: %d\n", p, c)
		}
		sb.WriteString("\n")
	}

	top := nc.TopLanguages(5)
	if len(top) > 0 {
		sb.WriteString("## Top языков\n")
		for _, s := range top {
			total := s.Manual + s.AI
			r := 0
			if total > 0 {
				r = int(float64(s.AI) * 100.0 / float64(total))
			}
			fmt.Fprintf(&sb, "- %s: %d chars total, %d%% AI\n", s.Lang, total, r)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Рекомендация\n")
	if ratio > 70 {
		sb.WriteString("AI-зависимость высокая (>70%). Попробуй раз в день написать функцию полностью руками — мышца не должна атрофироваться.\n")
	} else if ratio < 20 && totalSec > 0 {
		sb.WriteString("AI почти не используется. Если хочешь ускориться, попробуй Cursor или Copilot inline — для бойлерплейта это разгрузит руки.\n")
	} else {
		sb.WriteString("Баланс manual/AI в здоровом диапазоне. Продолжай.\n")
	}
	return sb.String()
}
