# ClickHouse performance — events_hourly_agg

Materialized view + SummingMergeTree target table для всех hot read queries.

## Architecture (cascading 3-tier)

```
   events (raw)            events_hourly_agg              events_daily_agg
   ────────────             ─────────────────              ─────────────────
   ts DateTime64(3)        bucket_ts DateTime           date Date
   user_id UUID    ──┐     user_id UUID         ──┐     user_id UUID
   category        ──│──→  category               │     category
   file_lang       ──│──→  file_lang             ─│──→  file_lang
   duration_ms     ──│──→  duration_ms (sum)      │     duration_ms (sum)
   chars_in        ──│──→  chars_in (sum)         │     chars_in (sum)
   lines_*         ──│──→  lines_* (sum)          │     lines_* (sum)
   ...               │     event_count (sum)      │     event_count (sum)
                     │                            │
              events_hourly_mv             events_daily_mv
                  GROUP BY                     GROUP BY
                  toStartOfHour(ts)            toDate(bucket_ts)
                  + user/cat/lang              + user/cat/lang
```

**Routing logic (store/clickhouse.go):**

| Query | since < 30d | since ≥ 30d |
|---|---|---|
| `AggregateByCategory` | hourly | **daily** |
| `AggregateByCategoryBulk` | hourly | **daily** |
| `DailyTrend` | hourly (tz-correct) | hourly |
| `LanguageBreakdown` | hourly | hourly |
| `Heatmap` | hourly (toDayOfWeek/Hour требует) | hourly |

- **events** остаётся raw для `ListRecent` (UI table) и forensics. TTL 18 мес.
- **events_hourly_agg** — `SummingMergeTree` агрегирует rows с одинаковым
  ORDER BY ключом при background merge'ах. Query всё равно делает `sum()`
  т.к. до merge'а могут быть partial rows.
- **events_hourly_mv** — катит каждый INSERT в events через GROUP BY и
  пишет в target. Только новые inserts; для backfill — отдельный INSERT INTO
  в migration 003.

## Reduction (cumulative)

100 active users × N rows/day:
- **events** (raw): ~10M/day (real prod scale)
- **events_hourly_agg**: 24 × 6 cat × ~20 lang × 100 users = ~290K/day → **35× vs raw**
- **events_daily_agg**: 1 × 6 × ~20 × 100 = ~12K/day → **24× vs hourly = 840× vs raw**

## Benchmark (10 users × 30 days × 10K events = 3M events)

Local CH 24, MacBook Pro M1, 1 thread. Query × 5 runs:

| Query                  | raw         | hourly       | daily          |
|------------------------|-------------|--------------|----------------|
| AggregateByCategory    | 4 ms median, 14 ms p95 | 1 ms / 8 ms p95 | **1 ms / 1 ms p95** ⭐ |
| LanguageBreakdown      | 6 ms median | 2 ms median  | **1 ms median** ⭐ |
| DailyTrend             | 5 ms median | 2 ms median  | (n/a — tz-sensitive) |
| Heatmap                | 4 ms median | 1 ms median  | (n/a — нужен hourly) |

**Daily MV** наиболее стабилен на p95 (~1ms константа vs hourly 8ms p95).
На 30M-events/day ожидается ещё больший gap — daily читает 30 dates × N users
× 6 cat × 20 langs ~ 3.6K rows vs hourly 86K rows.

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

## Phase 2: Redis cache (внутри store/cached.go)

`CachedEventStore` — decorator поверх EventStore, оборачивает CH-store если
Redis available (`cfg.RedisAddr` reachable). Per-method TTL'ы:

| Method | TTL | Why |
|---|---|---|
| `AggregateByCategory` | 10m | Insights / dashboard summary |
| `AggregateByCategoryBulk` | 5m | Team summary (multi-user, чаще меняется) |
| `LanguageBreakdown` | 10m | Languages widget, slow-changing |
| `DailyTrend` | 5m | Trend chart, active dashboards refresh |
| `Heatmap` | 10m | Heatmap widget, slow-changing |

Cache key scheme: `eop:cache:{prefix}:{userID}:{params}` — например
`eop:cache:agg:abc-123:1234567890`. Bulk variant использует sorted comma-
separated user list для key stability.

Pass-through (NOT cached): `Insert`, `ListRecent` (always-fresh), `ActiveUserIDs`
(admin), `DeleteUserData` (invalidates user's cached entries via SCAN).

Failure handling:
- Redis unreachable на startup → continuing without cache (logged)
- Redis fails mid-request → SET errors logged, GET returns miss (read
  fallthrough к inner store)
- 8/8 unit-тестов: hit/miss, per-user/since/tz keys, bulk sort stability,
  delete invalidation, broken-cache fallthrough

## Phase 3 (future): partitioning + projections

Если 30M+ events/день — ещё рассмотреть:
- **ClickHouse partitioning** по `customer_id` (или region) для multi-tenant
  isolation + per-tenant retention policies.
- **Projections on events_hourly_agg** для альтернативных sort orders (е.г.
  `ORDER BY (file_lang, bucket_ts)` для cross-user lang stats).
- **Per-team aggregate** для team-summary queries (сейчас идёт через
  AggregateByCategoryBulk над users в команде).
