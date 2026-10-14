-- The deploys table from the earlier exercises. Production has millions
-- of these rows; here 14 are enough to see what the query planner does.
CREATE TABLE deploys (
    id          INTEGER PRIMARY KEY,
    service     TEXT    NOT NULL,
    version     TEXT    NOT NULL,
    environment TEXT    NOT NULL,
    status      TEXT    NOT NULL,
    started_at  TEXT    NOT NULL, -- 'YYYY-MM-DD HH:MM:SS', UTC
    duration_s  INTEGER NOT NULL
);

-- Already there: an index on the start time.
CREATE INDEX deploys_by_start ON deploys (started_at);
