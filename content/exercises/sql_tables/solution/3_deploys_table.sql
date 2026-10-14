-- Task 3: Create the deploys table
--
-- Create a table named deploys with five columns:
--   id          INTEGER, the primary key
--   service_id  INTEGER, required, a foreign key to services (id)
--   version     TEXT, required
--   status      TEXT, required, 'running', 'succeeded' or 'failed'
--   started_at  TEXT, required
-- A service can't have two deploys of the same version.

CREATE TABLE deploys (
    id         INTEGER PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services (id),
    version    TEXT NOT NULL,
    status     TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
    started_at TEXT NOT NULL,
    UNIQUE (service_id, version)
);
