-- ClickHouse: события и attribution.

CREATE TABLE IF NOT EXISTS events (
    ts            DateTime64(3),
    user_id       UUID,
    device_id     UUID,
    session_id    UUID,
    app_bundle    LowCardinality(String),
    category      Enum8('idle' = 1, 'manual' = 2, 'ai' = 3, 'reading' = 4, 'refactor' = 5, 'other' = 6),
    source        Enum8('os' = 1, 'browser' = 2, 'ide' = 3, 'cli' = 4),
    ai_provider   LowCardinality(String),
    ai_channel    LowCardinality(String),
    project_id    UUID,
    file_lang     LowCardinality(String),
    duration_ms   UInt32,
    chars_in      UInt32,
    lines_added   UInt32,
    lines_removed UInt32,
    meta          String
) ENGINE = MergeTree
ORDER BY (user_id, ts)
PARTITION BY toYYYYMM(ts)
TTL toDateTime(ts) + INTERVAL 18 MONTH;

CREATE TABLE IF NOT EXISTS attribution_events (
    ts          DateTime64(3),
    user_id     UUID,
    project_id  UUID,
    file_lang   LowCardinality(String),
    category    Enum8('typed' = 1, 'pasted_ai' = 2, 'pasted_other' = 3, 'ai_inline' = 4, 'ai_agent' = 5, 'refactor' = 6, 'unknown' = 7),
    ai_provider LowCardinality(String),
    lines       UInt32,
    chars       UInt32,
    focus_ms    UInt32
) ENGINE = MergeTree
ORDER BY (user_id, ts)
PARTITION BY toYYYYMM(ts)
TTL toDateTime(ts) + INTERVAL 18 MONTH;
