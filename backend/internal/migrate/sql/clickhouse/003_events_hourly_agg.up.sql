-- Materialized aggregation для events за hourly buckets.
--
-- Покрывает hot queries:
--   AggregateByCategory      — sum(duration_ms) по category
--   AggregateByCategoryBulk  — то же по нескольким users
--   DailyTrend               — toDate(toTimeZone(bucket_ts, tz)) per day
--   LanguageBreakdown        — group by file_lang × category
--   Heatmap                  — toDayOfWeek/toHour(toTimeZone(bucket_ts, tz))
--
-- Hourly granularity сохраняется для tz-correct rollup'ов: aggregates за UTC
-- hour, а query-time переводим в user'ский tz через toTimeZone().
-- Reduction: ~330× для 100 active users (24 hr × 6 cat × ~20 lang = 2880
-- buckets/user/day vs миллионы raw events).
--
-- SummingMergeTree автоматически суммирует rows с одинаковым ORDER BY ключом
-- при background merge'ах. Query всё равно делает sum() т.к. до merge'а
-- могут быть partial rows.
--
-- ВАЖНО: MV catches только новые inserts с момента создания. Для backfill
-- pre-existing events нужно после применения migration выполнить вручную
-- (см. ниже после CREATE TABLE) — POPULATE несовместим с TO clause.
CREATE TABLE IF NOT EXISTS events_hourly_agg (
    bucket_ts     DateTime,
    user_id       UUID,
    category      Enum8('idle' = 1, 'manual' = 2, 'ai' = 3, 'reading' = 4, 'refactor' = 5, 'other' = 6),
    file_lang     LowCardinality(String),
    duration_ms   UInt64,
    chars_in      UInt64,
    lines_added   UInt64,
    lines_removed UInt64,
    event_count   UInt64
) ENGINE = SummingMergeTree((duration_ms, chars_in, lines_added, lines_removed, event_count))
ORDER BY (user_id, bucket_ts, category, file_lang)
PARTITION BY toYYYYMM(bucket_ts)
TTL bucket_ts + INTERVAL 18 MONTH;

CREATE MATERIALIZED VIEW IF NOT EXISTS events_hourly_mv TO events_hourly_agg AS
SELECT
    toStartOfHour(ts) AS bucket_ts,
    user_id,
    category,
    file_lang,
    toUInt64(sum(duration_ms))   AS duration_ms,
    toUInt64(sum(chars_in))      AS chars_in,
    toUInt64(sum(lines_added))   AS lines_added,
    toUInt64(sum(lines_removed)) AS lines_removed,
    count()                       AS event_count
FROM events
GROUP BY bucket_ts, user_id, category, file_lang;

-- Backfill для pre-existing events (если миграция применяется к не-пустой БД).
-- Безопасно double-count'ить события из последнего часа (которые попадут и в
-- MV) — мы insert'им только до hour preceding now.
INSERT INTO events_hourly_agg
SELECT
    toStartOfHour(ts) AS bucket_ts,
    user_id,
    category,
    file_lang,
    toUInt64(sum(duration_ms))   AS duration_ms,
    toUInt64(sum(chars_in))      AS chars_in,
    toUInt64(sum(lines_added))   AS lines_added,
    toUInt64(sum(lines_removed)) AS lines_removed,
    count()                       AS event_count
FROM events
WHERE ts < toStartOfHour(now()) - INTERVAL 1 HOUR
GROUP BY bucket_ts, user_id, category, file_lang;
