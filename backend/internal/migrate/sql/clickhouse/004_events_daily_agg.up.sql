-- Daily aggregate поверх events_hourly_agg — для long-range queries (30+ days).
--
-- Cascading MV chain:
--   events (raw)  →  events_hourly_agg  →  events_daily_agg
--
-- Granularity trade-off:
--   - Hourly MV: tz-correct query-time (toTimeZone), 24 buckets/user/day
--   - Daily MV: UTC-day, 1 bucket/user/day (~24× reduction поверх hourly)
--     На queries за 30+ дней tz-drift (~1h) не critical для общих агрегатов.
--
-- Reduction (cumulative):
--   raw → hourly: 35× (см. 003)
--   hourly → daily: 24×
--   raw → daily: 840× total
--
-- Routing:
--   - AggregateByCategory(since=last_30d+): читать daily
--   - AggregateByCategory(since=last_7d):    читать hourly (current)
--   - DailyTrend, LanguageBreakdown:         hourly (tz-sensitive)
--   - Heatmap:                                hourly (toDayOfWeek/toHour
--                                            требуют hourly granularity)

CREATE TABLE IF NOT EXISTS events_daily_agg (
    date          Date,
    user_id       UUID,
    category      Enum8('idle' = 1, 'manual' = 2, 'ai' = 3, 'reading' = 4, 'refactor' = 5, 'other' = 6),
    file_lang     LowCardinality(String),
    duration_ms   UInt64,
    chars_in      UInt64,
    lines_added   UInt64,
    lines_removed UInt64,
    event_count   UInt64
) ENGINE = SummingMergeTree((duration_ms, chars_in, lines_added, lines_removed, event_count))
ORDER BY (user_id, date, category, file_lang)
PARTITION BY toYYYYMM(date)
TTL toDateTime(date) + INTERVAL 24 MONTH;

CREATE MATERIALIZED VIEW IF NOT EXISTS events_daily_mv TO events_daily_agg AS
SELECT
    toDate(bucket_ts) AS date,
    user_id,
    category,
    file_lang,
    duration_ms,
    chars_in,
    lines_added,
    lines_removed,
    event_count
FROM events_hourly_agg;

-- Backfill для pre-existing hourly rows. Безопасно — если hourly уже backfilled
-- из 003, эта INSERT просуммирует те же chunks в daily-buckets через
-- SummingMergeTree merge.
INSERT INTO events_daily_agg
SELECT
    toDate(bucket_ts) AS date,
    user_id,
    category,
    file_lang,
    sum(duration_ms),
    sum(chars_in),
    sum(lines_added),
    sum(lines_removed),
    sum(event_count)
FROM events_hourly_agg
WHERE bucket_ts < toStartOfHour(now()) - INTERVAL 1 HOUR
GROUP BY date, user_id, category, file_lang;
