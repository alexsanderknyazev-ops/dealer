CREATE DATABASE IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.error_events
(
    occurred_at DateTime64(3, 'UTC'),
    event_id UUID,
    source LowCardinality(String),
    kind LowCardinality(String),
    severity LowCardinality(String),
    message String,
    error_code LowCardinality(String),
    http_status UInt16,
    grpc_code LowCardinality(String),
    trace_id String,
    request_id String,
    user_id String,
    route String,
    service LowCardinality(String),
    environment LowCardinality(String),
    context String,
    stack String,
    client String,
    ingested_at DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(occurred_at)
ORDER BY (service, occurred_at)
TTL occurred_at + INTERVAL 90 DAY;
