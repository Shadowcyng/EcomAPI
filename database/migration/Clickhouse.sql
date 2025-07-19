DROP TABLE IF EXISTS analytics_events;
CREATE TABLE analytics_events (
    event_id UUID,
    event_type String,
    user_id String,
    session_id String,
    timestamp DateTime64(3),
    page_path String,
    referrer String,
    user_agent String,
    ip_address String,
    duration_ms Int64,
    products String, -- To store json.RawMessage as a string
    location String, -- For timezone
    event_data JSON -- For flexible arbitrary data (JSON type requires ClickHouse v21.10+ or Cloud)
    -- If JSON type is not supported by your ClickHouse version, use String:
    -- event_data String
)
ENGINE = MergeTree()
ORDER BY (timestamp, event_type);




docker run -d --name clickhouse-server \
  --ulimit nofile=262144:262144 \
  -p 8123:8123 \
  -p 9000:9000 \
  -e CLICKHOUSE_USER=default \
  -e CLICKHOUSE_PASSWORD= \
  clickhouse/clickhouse-server