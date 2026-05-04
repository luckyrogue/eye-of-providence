package reports

import (
	"fmt"
	"strings"
)

// mockReport — детерминистичный отчёт для dev-режима, когда нет API key.
// Имитирует стиль реального Gemini-отчёта (эмодзи, дружелюбный тон).
func mockReport(nc *NumericContext) string {
	totalSec := nc.TotalActiveMS / 1000
	aiSec := nc.ByCategoryMS["ai"] / 1000
	manualSec := nc.ByCategoryMS["manual"] / 1000

	// Empty state — короткий, без формальной структуры
	if totalSec == 0 && nc.EventsTotal == 0 {
		return fmt.Sprintf(
			"# 📊 Отчёт за %s\n\n"+
				"На этой неделе тишина — данных нет 🌙. "+
				"Подключи агент или браузерное расширение, и в следующий раз будет что показать.\n\n"+
				"_Mock-режим: реальный Gemini не вызывался (нет API key)._\n",
			nc.Period,
		)
	}

	ratio := 0
	if totalSec > 0 {
		ratio = int(float64(aiSec) * 100.0 / float64(totalSec))
	}
	manualRatio := 100 - ratio

	var sb strings.Builder
	fmt.Fprintf(&sb, "# 📊 Отчёт за %s\n\n", nc.Period)
	sb.WriteString("_Mock-режим: реальный Gemini не вызывался (нет API key)._\n\n")

	// TL;DR
	sb.WriteString("## ⚡ TL;DR\n\n")
	if ratio >= 70 {
		fmt.Fprintf(&sb, "Ты сильно полагался на AI — **%d%%** времени с подсказками. Активного времени **%s**.\n\n", ratio, formatDuration(totalSec))
	} else if ratio >= 30 {
		fmt.Fprintf(&sb, "Здоровый баланс: **%d%% AI / %d%% вручную**. Всего активного времени **%s**.\n\n", ratio, manualRatio, formatDuration(totalSec))
	} else if ratio > 0 {
		fmt.Fprintf(&sb, "Преимущественно ручной код — только **%d%%** с AI. Активного времени **%s**.\n\n", ratio, formatDuration(totalSec))
	} else {
		fmt.Fprintf(&sb, "100%% ручной код за этот период. Активного времени **%s**.\n\n", formatDuration(totalSec))
	}

	// Ключевые цифры
	sb.WriteString("## 📈 Ключевые цифры\n\n")
	fmt.Fprintf(&sb, "- ⏱️ Активное время: **%s**\n", formatDuration(totalSec))
	if aiSec > 0 {
		fmt.Fprintf(&sb, "- 🧠 С AI: **%d%%** (%s)\n", ratio, formatDuration(aiSec))
	}
	if manualSec > 0 {
		fmt.Fprintf(&sb, "- ✍️ Вручную: **%d%%** (%s)\n", manualRatio, formatDuration(manualSec))
	}
	fmt.Fprintf(&sb, "- 📊 Событий обработано: **%d**\n", nc.EventsTotal)
	sb.WriteString("\n")

	// Провайдеры
	if len(nc.AICharsByProvider) > 0 {
		sb.WriteString("## 🤖 По провайдерам AI\n\n")
		for p, c := range nc.AICharsByProvider {
			if c > 0 {
				fmt.Fprintf(&sb, "- ▸ `%s` — **%d** символов\n", p, c)
			}
		}
		sb.WriteString("\n")
	}

	// Языки
	top := nc.TopLanguages(5)
	if len(top) > 0 {
		sb.WriteString("## 🎯 По языкам\n\n")
		for _, s := range top {
			total := s.Manual + s.AI
			r := 0
			if total > 0 {
				r = int(float64(s.AI) * 100.0 / float64(total))
			}
			fmt.Fprintf(&sb, "- `%s` — **%d** симв · AI %d%% / вручную %d%%\n", s.Lang, total, r, 100-r)
		}
		sb.WriteString("\n")
	}

	// Рекомендация
	sb.WriteString("## 🚀 Рекомендация\n\n")
	if ratio > 70 {
		sb.WriteString("AI-зависимость высокая. Попробуй раз в день написать функцию полностью руками — мышца не должна атрофироваться. ✨\n")
	} else if ratio < 20 && totalSec > 0 {
		sb.WriteString("AI почти не используется. Если хочешь ускориться, попробуй Cursor или Copilot inline — для бойлерплейта это разгрузит руки. ⚡\n")
	} else {
		sb.WriteString("Баланс manual/AI в здоровом диапазоне. Так держать! 🎯\n")
	}
	return sb.String()
}

func formatDuration(sec uint64) string {
	if sec < 60 {
		return fmt.Sprintf("%d сек", sec)
	}
	if sec < 3600 {
		return fmt.Sprintf("%d мин", sec/60)
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	if m == 0 {
		return fmt.Sprintf("%d ч", h)
	}
	return fmt.Sprintf("%d ч %d мин", h, m)
}
