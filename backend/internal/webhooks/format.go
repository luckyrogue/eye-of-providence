package webhooks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Format string

const (
	FormatRaw   Format = "raw"
	FormatSlack Format = "slack"
)

func validFormat(f string) bool {
	switch Format(f) {
	case FormatRaw, FormatSlack:
		return true
	}
	return false
}

func formatPayload(format Format, event string, payload any) ([]byte, error) {
	switch format {
	case FormatSlack:
		return marshalSlack(event, payload)
	case FormatRaw, "":
		return json.Marshal(map[string]any{
			"event":   event,
			"data":    payload,
			"sent_at": time.Now().UTC().Format(time.RFC3339),
		})
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

func marshalSlack(event string, payload any) ([]byte, error) {
	data, _ := payload.(map[string]any)

	var headerText, summaryText string
	fields := []map[string]any{}

	switch event {
	case EventCommitIngested:
		sha, _ := data["sha"].(string)
		message, _ := data["message"].(string)
		branch, _ := data["branch"].(string)
		linesAdded := numField(data, "lines_added")
		linesRemoved := numField(data, "lines_removed")

		shortSHA := sha
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}
		headerText = fmt.Sprintf(":rocket: New commit %s", shortSHA)
		summaryText = truncate(firstLine(message), 80)
		if branch != "" {
			fields = append(fields, slackField("Branch", "`"+branch+"`"))
		}
		if linesAdded > 0 || linesRemoved > 0 {
			fields = append(fields, slackField("Changes", fmt.Sprintf("+%d / -%d", linesAdded, linesRemoved)))
		}

	case EventReportGenerated:
		period, _ := data["period"].(string)
		model, _ := data["model"].(string)
		headerText = ":bar_chart: New report"
		summaryText = period
		if model != "" {
			fields = append(fields, slackField("Model", "`"+model+"`"))
		}
		if period != "" {
			fields = append(fields, slackField("Period", period))
		}

	case EventAnomalyDetected:
		kind, _ := data["kind"].(string)
		category, _ := data["category"].(string)
		z := numField(data, "z_score")
		yesterdayMS := numField(data, "yesterday_ms")
		baselineMS := numField(data, "baseline_ms")
		emoji, label := anomalyHeader(kind)
		headerText = emoji + " " + label
		yHrs := float64(yesterdayMS) / 3600000
		bHrs := float64(baselineMS) / 3600000
		summaryText = fmt.Sprintf("Yesterday: %.1fh · baseline: %.1fh", yHrs, bHrs)
		if category != "" {
			fields = append(fields, slackField("Category", "`"+category+"`"))
		}
		fields = append(fields, slackField("Z-score", fmt.Sprintf("%.2f", float64(z)/1)))

	default:

		headerText = ":bell: Eye of Providence event"
		summaryText = event
	}

	blocks := []map[string]any{
		{
			"type": "header",
			"text": map[string]any{"type": "plain_text", "text": headerText},
		},
	}
	if summaryText != "" {
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{"type": "mrkdwn", "text": summaryText},
		})
	}
	if len(fields) > 0 {
		blocks = append(blocks, map[string]any{
			"type":   "section",
			"fields": fields,
		})
	}

	return json.Marshal(map[string]any{

		"text":   strings.TrimSpace(headerText + " — " + summaryText),
		"blocks": blocks,
	})
}

func anomalyHeader(kind string) (emoji, label string) {
	switch kind {
	case "ai_high":
		return ":rocket:", "Unusually high AI usage"
	case "ai_low":
		return ":turtle:", "AI usage dipped"
	case "manual_high":
		return ":hammer_and_wrench:", "High manual coding day"
	case "manual_low":
		return ":zzz:", "Manual coding dipped"
	case "refactor_high":
		return ":sparkles:", "Lots of refactoring"
	case "activity_high":
		return ":chart_with_upwards_trend:", "Productivity spike"
	case "activity_low":
		return ":warning:", "Activity dropped"
	}
	return ":bell:", "Anomaly detected"
}

func slackField(label, value string) map[string]any {
	return map[string]any{
		"type": "mrkdwn",
		"text": fmt.Sprintf("*%s*\n%s", label, value),
	}
}

func numField(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func truncate(s string, n int) string {
	if len([]rune(s)) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n-1]) + "…"
}
