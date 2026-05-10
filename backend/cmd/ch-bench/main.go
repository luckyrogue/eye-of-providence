// ch-bench — synthesizes events, runs hot queries в две версии (events
// table vs events_hourly_agg MV) и сравнивает latency.
//
// Usage:
//
//	EOP_CH_DSN="clickhouse://default:@localhost:9000/eop_test" \
//	  go run ./cmd/ch-bench --users=10 --days=30 --events-per-user-per-day=1000
//
// Default: 10 users × 30 days × 1000 events = 300K events, ~5min на CH Cloud.
// На локальной CH (docker) обычно <30s. Larger sample (--events-per-user-per-day=10000)
// для real 10M-в-день simulation.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	mathrand "math/rand"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/migrate"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	users := flag.Int("users", 10, "synthetic users")
	days := flag.Int("days", 30, "days of history")
	perDay := flag.Int("events-per-user-per-day", 1000, "events/user/day")
	skipSeed := flag.Bool("skip-seed", false, "пропустить генерацию (если БД уже заполнена)")
	flag.Parse()

	dsn := os.Getenv("EOP_CH_DSN")
	if dsn == "" {
		return fmt.Errorf("EOP_CH_DSN required")
	}

	if err := migrate.RunClickHouse(context.Background(), dsn); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return err
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return err
	}
	defer conn.Close()

	totalEvents := *users * *days * *perDay
	fmt.Printf("=== ch-bench ===\nusers=%d days=%d perDay=%d total=%d\n", *users, *days, *perDay, totalEvents)

	uids := make([]uuid.UUID, *users)
	for i := range uids {
		uids[i] = uuid.New()
	}

	if !*skipSeed {
		if err := seed(conn, uids, *days, *perDay); err != nil {
			return fmt.Errorf("seed: %w", err)
		}
		// Force MV catch-up + flush merges. SummingMergeTree merge'ит в
		// background; OPTIMIZE FINAL ускоряет для бенчмарка чтобы измерять
		// merged-state latency.
		fmt.Println("OPTIMIZE FINAL events_hourly_agg ...")
		if err := conn.Exec(context.Background(), "OPTIMIZE TABLE events_hourly_agg FINAL"); err != nil {
			log.Printf("optimize hourly: %v", err)
		}
		fmt.Println("OPTIMIZE FINAL events_daily_agg ...")
		if err := conn.Exec(context.Background(), "OPTIMIZE TABLE events_daily_agg FINAL"); err != nil {
			log.Printf("optimize daily: %v", err)
		}
	}

	uid := uids[0]
	since := time.Now().Add(-time.Duration(*days) * 24 * time.Hour)

	cases := []struct {
		name  string
		query string
	}{
		{"AggregateByCategory_raw",
			`SELECT category, sum(duration_ms) FROM events WHERE user_id = ? AND ts >= ? GROUP BY category`},
		{"AggregateByCategory_hourly",
			`SELECT category, sum(duration_ms) FROM events_hourly_agg WHERE user_id = ? AND bucket_ts >= ? GROUP BY category`},
		{"AggregateByCategory_daily",
			`SELECT category, sum(duration_ms) FROM events_daily_agg WHERE user_id = ? AND date >= ? GROUP BY category`},
		{"DailyTrend_raw",
			`SELECT toDate(toTimeZone(ts, 'UTC')) AS d, category, sum(chars_in), sum(duration_ms) FROM events WHERE user_id = ? AND ts >= ? GROUP BY d, category`},
		{"DailyTrend_hourly",
			`SELECT toDate(toTimeZone(bucket_ts, 'UTC')) AS d, category, sum(chars_in), sum(duration_ms) FROM events_hourly_agg WHERE user_id = ? AND bucket_ts >= ? GROUP BY d, category`},
		{"LanguageBreakdown_raw",
			`SELECT file_lang, category, sum(chars_in), sum(duration_ms) FROM events WHERE user_id = ? AND ts >= ? AND file_lang != '' GROUP BY file_lang, category`},
		{"LanguageBreakdown_hourly",
			`SELECT file_lang, category, sum(chars_in), sum(duration_ms) FROM events_hourly_agg WHERE user_id = ? AND bucket_ts >= ? AND file_lang != '' GROUP BY file_lang, category`},
		{"LanguageBreakdown_daily",
			`SELECT file_lang, category, sum(chars_in), sum(duration_ms) FROM events_daily_agg WHERE user_id = ? AND date >= ? AND file_lang != '' GROUP BY file_lang, category`},
		{"Heatmap_raw",
			`SELECT toDayOfWeek(toTimeZone(ts, 'UTC')) AS d, toHour(toTimeZone(ts, 'UTC')) AS h, category, sum(duration_ms) FROM events WHERE user_id = ? AND ts >= ? GROUP BY d, h, category`},
		{"Heatmap_hourly",
			`SELECT toDayOfWeek(toTimeZone(bucket_ts, 'UTC')) AS d, toHour(toTimeZone(bucket_ts, 'UTC')) AS h, category, sum(duration_ms) FROM events_hourly_agg WHERE user_id = ? AND bucket_ts >= ? GROUP BY d, h, category`},
	}

	fmt.Println("\nquery                            min       median    p95       runs")
	fmt.Println("--------------------------------------------------------------------")
	for _, c := range cases {
		latencies := bench(conn, c.query, uid, since, 5)
		fmt.Printf("%-32s %-9s %-9s %-9s %d\n",
			c.name,
			fmtDur(latencies[0]),
			fmtDur(latencies[len(latencies)/2]),
			fmtDur(latencies[len(latencies)-1]),
			len(latencies))
	}
	return nil
}

func seed(conn driver.Conn, uids []uuid.UUID, days, perDay int) error {
	ctx := context.Background()
	apps := []string{"com.microsoft.VSCode", "com.todesktop.230313mzl4w4u92", "Google Chrome"}
	categories := []string{"manual", "ai", "refactor", "reading", "other"}
	langs := []string{"go", "typescript", "python", "rust", "javascript", ""}

	rng := mathrand.New(mathrand.NewSource(42))
	now := time.Now().UTC()

	for di := 0; di < days; di++ {
		dayStart := now.Add(-time.Duration(di+1) * 24 * time.Hour)
		batch, err := conn.PrepareBatch(ctx, `
			INSERT INTO events (
				ts, user_id, device_id, session_id, app_bundle, category, source,
				ai_provider, ai_channel, project_id, file_lang,
				duration_ms, chars_in, lines_added, lines_removed, meta
			)`)
		if err != nil {
			return err
		}
		for _, uid := range uids {
			for ei := 0; ei < perDay; ei++ {
				ts := dayStart.Add(time.Duration(rng.Int63n(int64(24 * time.Hour))))
				app := apps[rng.Intn(len(apps))]
				cat := categories[rng.Intn(len(categories))]
				lang := langs[rng.Intn(len(langs))]
				if err := batch.Append(
					ts, uid, uuid.Nil, uuid.Nil,
					app, cat, "ide",
					"", "", uuid.Nil, lang,
					uint32(rng.Intn(60_000)), uint32(rng.Intn(500)), uint32(rng.Intn(20)), uint32(rng.Intn(10)), "",
				); err != nil {
					return err
				}
			}
		}
		if err := batch.Send(); err != nil {
			return err
		}
		fmt.Printf("seeded day %d/%d\n", di+1, days)
	}
	return nil
}

// bench — runs query 5 раз, returns sorted latencies в asc.
func bench(conn driver.Conn, query string, uid uuid.UUID, since time.Time, runs int) []time.Duration {
	out := make([]time.Duration, runs)
	for i := 0; i < runs; i++ {
		start := time.Now()
		rows, err := conn.Query(context.Background(), query, uid, since)
		if err != nil {
			log.Printf("query failed: %v", err)
			out[i] = -1
			continue
		}
		for rows.Next() {
			// drain
		}
		if err := rows.Close(); err != nil {
			log.Printf("close rows: %v", err)
			out[i] = -1
			continue
		}
		out[i] = time.Since(start)
	}
	// insertion sort — runs малое число
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

func fmtDur(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
