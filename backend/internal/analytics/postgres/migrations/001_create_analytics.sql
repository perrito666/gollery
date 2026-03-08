-- Analytics events table: records individual view/click events.
-- Privacy-safe: stores hashed identifiers, not raw IPs or user agents.
CREATE TABLE IF NOT EXISTS analytics_events (
    id          BIGSERIAL PRIMARY KEY,
    event_type  TEXT        NOT NULL,
    object_id   TEXT        NOT NULL,
    visitor_hash TEXT       NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_object_id ON analytics_events (object_id);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON analytics_events (created_at);
CREATE INDEX IF NOT EXISTS idx_events_type_object ON analytics_events (event_type, object_id);

CREATE TABLE IF NOT EXISTS popularity_daily (
    object_id   TEXT        NOT NULL,
    event_type  TEXT        NOT NULL,
    day         DATE        NOT NULL,
    view_count  BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (object_id, event_type, day)
);

CREATE INDEX IF NOT EXISTS idx_popularity_day ON popularity_daily (day);

---- create above / drop below ----

DROP TABLE IF EXISTS popularity_daily;
DROP TABLE IF EXISTS analytics_events;
