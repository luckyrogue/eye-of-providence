package metrics

import (
	"strings"
	"testing"
	"time"
)

func TestHistogram_Observe_BucketsCumulative(t *testing.T) {
	h := newHistogram("test_h", "test")

	h.Observe(3 * time.Millisecond)  // → 5ms, 10ms, 25ms, ... все верхние bucket'ы
	h.Observe(50 * time.Millisecond) // → 50ms, 100ms, 250ms, ...
	h.Observe(2 * time.Second)       // → 2.5s, 5s, 10s

	if h.totalCount.Load() != 3 {
		t.Errorf("totalCount=%d, want 3", h.totalCount.Load())
	}

	// 3ms падает в КАЖДЫЙ bucket (все ≥ 3ms).
	for i, b := range h.buckets {
		got := h.bucketCounts[i].Load()
		want := uint64(0)
		if 3*time.Millisecond <= b {
			want++
		}
		if 50*time.Millisecond <= b {
			want++
		}
		if 2*time.Second <= b {
			want++
		}
		if got != want {
			t.Errorf("bucket %v: count=%d, want %d", b, got, want)
		}
	}
}

func TestHistogram_NegativeIgnored(t *testing.T) {
	h := newHistogram("neg", "test")
	h.Observe(-1 * time.Millisecond)
	if h.totalCount.Load() != 0 {
		t.Error("negative observation accepted")
	}
}

func TestRender_IncludesCountersAndHistograms(t *testing.T) {
	IngestEventsAccepted.Add(42)
	RequestLatency.Observe(15 * time.Millisecond)

	out := Render()

	mustContain := []string{
		"# HELP eop_ingest_events_accepted_total",
		"# TYPE eop_ingest_events_accepted_total counter",
		"eop_ingest_events_accepted_total ",

		"# HELP eop_http_request_duration_seconds",
		"# TYPE eop_http_request_duration_seconds histogram",
		`eop_http_request_duration_seconds_bucket{le="0.005"}`,
		`eop_http_request_duration_seconds_bucket{le="+Inf"}`,
		"eop_http_request_duration_seconds_sum",
		"eop_http_request_duration_seconds_count",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("Render() missing %q", s)
		}
	}
}

func TestSnapshot_IncludesHistogramSummaries(t *testing.T) {
	ClickHouseWrite.Observe(20 * time.Millisecond)
	snap := Snapshot()
	if _, ok := snap["eop_clickhouse_write_duration_seconds_count"]; !ok {
		t.Error("snapshot missing histogram count")
	}
	if _, ok := snap["eop_clickhouse_write_duration_seconds_sum_ms"]; !ok {
		t.Error("snapshot missing histogram sum_ms")
	}
}
