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
TTL toDateTime(ts) + INTERVAL 18 MONTH
