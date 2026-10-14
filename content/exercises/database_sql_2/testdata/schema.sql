CREATE TABLE services (
    name            TEXT PRIMARY KEY,
    current_version TEXT NOT NULL -- what production runs right now
);

CREATE TABLE deploys (
    id          INTEGER PRIMARY KEY,
    service     TEXT NOT NULL,
    version     TEXT NOT NULL,
    environment TEXT NOT NULL, -- 'staging' or 'production'
    status      TEXT NOT NULL, -- 'running', 'succeeded', 'failed' or 'rolled_back'
    started_at  TEXT NOT NULL, -- 'YYYY-MM-DD HH:MM:SS', UTC
    finished_at TEXT           -- NULL while the deploy is still running
);
