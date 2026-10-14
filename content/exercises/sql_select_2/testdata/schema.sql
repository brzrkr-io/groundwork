-- One row per deploy: which version of which service went to which
-- environment, how it ended, when it started and how long it took.
CREATE TABLE deploys (
    id          INTEGER PRIMARY KEY,
    service     TEXT    NOT NULL,
    version     TEXT    NOT NULL,
    environment TEXT    NOT NULL, -- 'staging' or 'production'
    status      TEXT    NOT NULL, -- 'succeeded', 'failed' or 'rolled_back'
    started_at  TEXT    NOT NULL, -- 'YYYY-MM-DD HH:MM:SS', UTC
    duration_s  INTEGER NOT NULL  -- seconds from start to finish
);
