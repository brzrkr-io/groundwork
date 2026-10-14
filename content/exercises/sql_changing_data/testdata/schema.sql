CREATE TABLE services (
    id              INTEGER PRIMARY KEY,
    name            TEXT NOT NULL UNIQUE,
    current_version TEXT NOT NULL -- what production runs right now
);

CREATE TABLE deploys (
    id          INTEGER PRIMARY KEY,
    service_id  INTEGER NOT NULL REFERENCES services (id),
    version     TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('staging', 'production')),
    status      TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed', 'rolled_back')),
    started_at  TEXT NOT NULL -- 'YYYY-MM-DD HH:MM:SS', UTC
);
