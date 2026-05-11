// Package metrics — минимальный Prometheus-совместимый text exposition format
// без внешних зависимостей. Достаточно для Phase 7+; реальный Prometheus client
// можно подключить позже без изменения публичного API.
//
// Counters: ингест-события, ошибки, gemini calls (input/output tokens).
// Histograms: request latency, ClickHouse insert/query latency.
package metrics

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Counter struct {
	name string
	help string
	val  atomic.Uint64
}

func (c *Counter) Inc()          { c.val.Add(1) }
func (c *Counter) Add(v uint64)  { c.val.Add(v) }
func (c *Counter) Value() uint64 { return c.val.Load() }

//
// Bucketed observe в наносекундах для точности; рендерится в seconds-format
// (Prometheus convention).
//
// Buckets: 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s.
// +Inf bucket — сумма всех наблюдений.

var defaultBuckets = []time.Duration{
	5 * time.Millisecond,
	10 * time.Millisecond,
	25 * time.Millisecond,
	50 * time.Millisecond,
	100 * time.Millisecond,
	250 * time.Millisecond,
	500 * time.Millisecond,
	1 * time.Second,
	2500 * time.Millisecond,
	5 * time.Second,
	10 * time.Second,
}

type Histogram struct {
	name    string
	help    string
	buckets []time.Duration
	// Один atomic counter на каждый bucket (cumulative — observation падает
	// в каждый bucket >= её значения).
	bucketCounts []atomic.Uint64
	totalCount   atomic.Uint64
	sumNS        atomic.Uint64
}

func newHistogram(name, help string) *Histogram {
	return &Histogram{
		name:         name,
		help:         help,
		buckets:      defaultBuckets,
		bucketCounts: make([]atomic.Uint64, len(defaultBuckets)),
	}
}

func (h *Histogram) Observe(d time.Duration) {
	if d < 0 {
		return
	}
	h.totalCount.Add(1)
	h.sumNS.Add(uint64(d))
	// Cumulative buckets: каждое наблюдение падает в каждый bucket с le ≥ d.
	for i, b := range h.buckets {
		if d <= b {
			h.bucketCounts[i].Add(1)
		}
	}
}

// ObserveSince — удобная обёртка для defer метрик: defer h.ObserveSince(start).
func (h *Histogram) ObserveSince(start time.Time) {
	h.Observe(time.Since(start))
}

var (
	IngestEventsAccepted = &Counter{name: "eop_ingest_events_accepted_total", help: "Events accepted by /v1/ingest"}
	IngestEventsRejected = &Counter{name: "eop_ingest_events_rejected_total", help: "Events rejected by /v1/ingest"}
	IngestErrors         = &Counter{name: "eop_ingest_errors_total", help: "Errors during /v1/ingest insert"}

	ReportsGenerated   = &Counter{name: "eop_reports_generated_total", help: "Reports successfully generated (mock + gemini)"}
	GeminiCallsTotal   = &Counter{name: "eop_gemini_calls_total", help: "Real Gemini API calls (excluding mock)"}
	GeminiInputTokens  = &Counter{name: "eop_gemini_input_tokens_total", help: "Approx input tokens sent to Gemini"}
	GeminiOutputTokens = &Counter{name: "eop_gemini_output_tokens_total", help: "Approx output tokens received from Gemini"}

	UsersDeleted = &Counter{name: "eop_users_deleted_total", help: "Users that triggered DELETE /v1/me/data"}
)

var (
	RequestLatency  = newHistogram("eop_http_request_duration_seconds", "HTTP request latency by Fiber middleware")
	ClickHouseWrite = newHistogram("eop_clickhouse_write_duration_seconds", "ClickHouse Insert/Bulk write latency")
	ClickHouseRead  = newHistogram("eop_clickhouse_read_duration_seconds", "ClickHouse Query (aggregate/list) latency")
)

func allCounters() []*Counter {
	return []*Counter{
		IngestEventsAccepted, IngestEventsRejected, IngestErrors,
		ReportsGenerated, GeminiCallsTotal, GeminiInputTokens, GeminiOutputTokens,
		UsersDeleted,
	}
}

func allHistograms() []*Histogram {
	return []*Histogram{RequestLatency, ClickHouseWrite, ClickHouseRead}
}

// renderMu — защищаем от concurrent чтения histogram'ов. Каждый bucket atomic,
// но чтобы snapshot был согласованным (count соответствует bucket-counts),
// берём mutex.
var renderMu sync.Mutex

func Render() string {
	renderMu.Lock()
	defer renderMu.Unlock()
	var sb strings.Builder

	for _, c := range allCounters() {
		fmt.Fprintf(&sb, "# HELP %s %s\n", c.name, c.help)
		fmt.Fprintf(&sb, "# TYPE %s counter\n", c.name)
		fmt.Fprintf(&sb, "%s %d\n", c.name, c.Value())
	}

	for _, h := range allHistograms() {
		fmt.Fprintf(&sb, "# HELP %s %s\n", h.name, h.help)
		fmt.Fprintf(&sb, "# TYPE %s histogram\n", h.name)
		for i, b := range h.buckets {
			fmt.Fprintf(&sb, "%s_bucket{le=\"%g\"} %d\n", h.name, b.Seconds(), h.bucketCounts[i].Load())
		}
		total := h.totalCount.Load()
		// +Inf bucket — total count.
		fmt.Fprintf(&sb, "%s_bucket{le=\"+Inf\"} %d\n", h.name, total)
		fmt.Fprintf(&sb, "%s_sum %g\n", h.name, time.Duration(h.sumNS.Load()).Seconds())
		fmt.Fprintf(&sb, "%s_count %d\n", h.name, total)
	}

	return sb.String()
}

// Snapshot — JSON-friendly карта для /v1/admin/cost. Histogram'ы экспонируем
// как count + sum_ms (детальные buckets — только в Render для Prometheus).
func Snapshot() map[string]uint64 {
	out := make(map[string]uint64, len(allCounters())+len(allHistograms())*2)
	for _, c := range allCounters() {
		out[c.name] = c.Value()
	}
	for _, h := range allHistograms() {
		out[h.name+"_count"] = h.totalCount.Load()
		out[h.name+"_sum_ms"] = uint64(time.Duration(h.sumNS.Load()).Milliseconds())
	}
	return out
}
