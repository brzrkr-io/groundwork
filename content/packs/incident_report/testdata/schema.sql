CREATE TABLE teams (
    id            INTEGER PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    slack_channel TEXT NOT NULL
);

CREATE TABLE services (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    team_id INTEGER NOT NULL REFERENCES teams (id)
);

CREATE TABLE incidents (
    id                 INTEGER PRIMARY KEY,
    service_id         INTEGER NOT NULL REFERENCES services (id),
    severity           TEXT NOT NULL CHECK (severity IN ('sev1', 'sev2', 'sev3')), -- sev1 is the worst
    title              TEXT NOT NULL,
    opened_at          TEXT NOT NULL, -- 'YYYY-MM-DD HH:MM:SS', UTC
    resolved_at        TEXT,          -- NULL while the incident is open
    minutes_to_resolve INTEGER        -- NULL while the incident is open
);

-- A written review of an incident. At most one per incident.
CREATE TABLE postmortems (
    incident_id  INTEGER PRIMARY KEY REFERENCES incidents (id),
    author       TEXT NOT NULL,
    published_at TEXT NOT NULL
);
