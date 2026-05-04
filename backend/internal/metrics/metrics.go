// Package metrics — минимальный Prometheus-совместимый text exposition format
// без внешних зависимостей. Достаточно для Phase 7+; реальный Prometheus client
// можно подключить позже без изменения публичного API.
//
// Counters: ингест-события, ошибки, gemini calls (input/output tokens).
package metrics

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type Counter struct {
	name string
	help string
	val  atomic.Uint64
}

func (c *Counter) Inc()           { c.val.Add(1) }
func (c *Counter) Add(v uint64)   { c.val.Add(v) }
func (c *Counter) Value() uint64  { return c.val.Load() }

var (
	IngestEventsAccepted = &Counter{name: "eop_ingest_events_accepted_total", help: "Events accepted by /v1/ingest"}
	IngestEventsRejected = &Counter{name: "eop_ingest_events_rejected_total", help: "Events rejected by /v1/ingest"}
	IngestErrors         = &Counter{name: "eop_ingest_errors_total", help: "Errors during /v1/ingest insert"}

	ReportsGenerated     = &Counter{name: "eop_reports_generated_total", help: "Reports successfully generated (mock + gemini)"}
	GeminiCallsTotal     = &Counter{name: "eop_gemini_calls_total", help: "Real Gemini API calls (excluding mock)"}
	GeminiInputTokens    = &Counter{name: "eop_gemini_input_tokens_total", help: "Approx input tokens sent to Gemini"}
	GeminiOutputTokens   = &Counter{name: "eop_gemini_output_tokens_total", help: "Approx output tokens received from Gemini"}

	UsersDeleted = &Counter{name: "eop_users_deleted_total", help: "Users that triggered DELETE /v1/me/data"}
)

func all() []*Counter {
	return []*Counter{
		IngestEventsAccepted, IngestEventsRejected, IngestErrors,
		ReportsGenerated, GeminiCallsTotal, GeminiInputTokens, GeminiOutputTokens,
		UsersDeleted,
	}
}

// Render — Prometheus text exposition format.
func Render() string {
	var sb strings.Builder
	for _, c := range all() {
		fmt.Fprintf(&sb, "# HELP %s %s\n", c.name, c.help)
		fmt.Fprintf(&sb, "# TYPE %s counter\n", c.name)
		fmt.Fprintf(&sb, "%s %d\n", c.name, c.Value())
	}
	return sb.String()
}

// Snapshot — JSON-friendly карта для /v1/admin/cost.
func Snapshot() map[string]uint64 {
	out := make(map[string]uint64, len(all()))
	for _, c := range all() {
		out[c.name] = c.Value()
	}
	return out
}
