-- Teams own services. A service has one team (or none yet), a team can
-- own many services.
CREATE TABLE teams (
    id            INTEGER PRIMARY KEY,
    name          TEXT NOT NULL,
    slack_channel TEXT NOT NULL
);

CREATE TABLE services (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL,
    team_id INTEGER REFERENCES teams (id), -- empty (NULL) when nobody owns it yet
    tier    TEXT NOT NULL                  -- 'critical' or 'standard'
);

-- Who carries the pager for a service this week. A service with nobody on
-- call has no row here.
CREATE TABLE oncall (
    service_id INTEGER PRIMARY KEY REFERENCES services (id),
    engineer   TEXT NOT NULL
);

-- One row per deploy, pointing at its service by id.
CREATE TABLE deploys (
    id          INTEGER PRIMARY KEY,
    service_id  INTEGER NOT NULL REFERENCES services (id),
    version     TEXT NOT NULL,
    environment TEXT NOT NULL, -- 'staging' or 'production'
    status      TEXT NOT NULL, -- 'succeeded', 'failed' or 'rolled_back'
    started_at  TEXT NOT NULL  -- 'YYYY-MM-DD HH:MM:SS', UTC
);
