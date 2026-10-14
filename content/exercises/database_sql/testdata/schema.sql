CREATE TABLE deploys (
    id          INTEGER PRIMARY KEY,
    service     TEXT NOT NULL,
    version     TEXT NOT NULL,
    environment TEXT NOT NULL, -- 'staging' or 'production'
    status      TEXT NOT NULL, -- 'running', 'succeeded', 'failed' or 'rolled_back'
    started_at  TEXT NOT NULL  -- 'YYYY-MM-DD HH:MM:SS', UTC
);
