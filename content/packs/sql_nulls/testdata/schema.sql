-- Deploys recorded by two generations of deploy scripts. The old script
-- didn't record everything, and a running deploy has no end yet, so some
-- columns may be empty (NULL).
CREATE TABLE deploys (
    id          INTEGER PRIMARY KEY,
    service     TEXT NOT NULL,
    version     TEXT NOT NULL,
    environment TEXT,          -- 'staging' or 'production'; NULL: not recorded
    status      TEXT NOT NULL, -- 'running', 'succeeded' or 'failed'
    started_at  TEXT NOT NULL, -- 'YYYY-MM-DD HH:MM:SS', UTC
    finished_at TEXT,          -- NULL while the deploy is still running
    duration_s  INTEGER,       -- NULL while running, or when not recorded
    deployed_by TEXT           -- NULL when the CI pipeline deployed it
);
