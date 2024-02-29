CREATE TABLE IF NOT EXISTS processed_events (
    id SERIAL PRIMARY KEY,
    sent_at timestamp with time zone NOT NULL,
    event_type TEXT NOT NULL DEFAULT '',
    count INTEGER NOT NULL DEFAULT 1,
    UNIQUE (sent_at, event_type)
);
