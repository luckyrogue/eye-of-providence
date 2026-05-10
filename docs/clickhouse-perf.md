# ClickHouse performance — events_hourly_agg

Materialized view + SummingMergeTree target table для всех hot read queries.

## Architecture

```
              events (raw)                 events_hourly_agg
              ──────────────              ────────────────────
              ts DateTime64(3)            bucket_ts DateTime
              user_id UUID         ──┐    user_id UUID
              category Enum8         │    category Enum8
              file_lang LCS          │    file_lang LCS
              duration_ms UInt32     │    duration_ms UInt64    ← sum
              chars_in UInt32        ├──→ chars_in UInt64       ← sum
              lines_added UInt32     │    lines_added UInt64    ← sum
              lines_removed UInt32   │    lines_removed UInt64  ← sum
              ...                    │    event_count UInt64    ← count
                                     │
                              events_hourly_mv
                              (MATERIALIZED VIEW)
                                  GROUP BY (
                                    toStartOfHour(ts),
                                    user_id, category, file_lang
                                  )
```

- **events** остаётся raw для `ListRecent` (UI table) и forensics. TTL 18 мес.
- **events_hourly_agg** — `SummingMergeTree` агрегирует rows с одинаковым
  ORDER BY ключом при background merge'ах. Query всё равно делает `sum()`
  т.к. до merge'а могут быть partial rows.
- **events_hourly_mv** — катит каждый INSERT в events через GROUP BY и
  пишет в target. Только новые inserts; для backfill — отдельный INSERT INTO
  в migration 003.

## Reduction

100 active users × 24 hours × 6 categories × ~20 langs = ~290K rows/day в MV
vs 10M raw events/day = **~35× compression** при равной precision (часовая
granularity для всех queries — DailyTrend/Heatmap/Lang/Aggregate).

## Benchmark (10 users × 30 days × 10K events = 3M events)

Local CH 24, MacBook Pro M1, 1 thread. Query × 5 runs:

| Query                  | events (raw) | events_hourly_agg (MV) | Speedup |
|------------------------|--------------|------------------------|---------|
| AggregateByCategory    | 8 ms median  | 3 ms median            | 2.7×    |
| DailyTrend             | 7 ms median  | 2 ms median            | 3.5×    |
| LanguageBreakdown      | 7 ms median  | 2 ms median            | 3.5×    |
| Heatmap                | 7 ms median  | 1 ms median            | 7×      |

P95 latency reduction ещё сильнее: 18-20 ms → 2-7 ms (3-10×). На 10M-events/day
ожидается линейный рост raw query, MV остаётся ~constant в hourly buckets.

## Run benchmark

```sh
docker run -d --rm --name eop-ch -p 18123:8123 -p 19000:9000 clickhouse/clickhouse-server:24
docker exec eop-ch clickhouse-client --query="CREATE DATABASE eop_test"

EOP_CH_DSN="clickhouse://default:@localhost:19000/eop_test" \
  go run ./cmd/ch-bench --users=10 --days=30 --events-per-user-per-day=10000
```

Flags: `--skip-seed` если БД уже заполнена. Default: 10×30×1000 = 300K
(быстро, ~30s); 10×30×10000 = 3M (~3 мин); 100×30×10000 = 30M (~30 мин,
real prod scale).

## When to backfill

`migrations/clickhouse/003_events_hourly_agg.up.sql` делает backfill
автоматически при первом применении. Если migration применяется к prod-БД с
большим объёмом, monitor `system.merges` чтобы убедиться что backfill insert
не залочил background merges. На billions-rows tables — лучше manual chunked
backfill (по дням / week ranges).

## Future optimization

Если 30M events/день — ещё рассмотреть:
- **Daily aggregate table** (events_daily_agg) поверх hourly: `INSERT INTO
  events_daily_agg SELECT toDate(bucket_ts), ...` для queries с `days=30+`.
  Дополнительная редукция ~24×.
- **Per-team aggregate** для team-summary queries (сейчас идёт через
  AggregateByCategoryBulk над users в команде).
- **Projections on events_hourly_agg** для альтернативных sort orders (е.г.
  `ORDER BY (file_lang, bucket_ts)` для cross-user lang stats).
