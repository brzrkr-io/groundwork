-- A finished services table with a few rows, for the deploys and
-- migration tasks. It needs testdata/teams.sql first.
CREATE TABLE services (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    team_id INTEGER NOT NULL REFERENCES teams (id),
    tier    TEXT NOT NULL DEFAULT 'standard' CHECK (tier IN ('critical', 'standard'))
);

INSERT INTO services (id, name, team_id, tier) VALUES
    (1, 'api',      2, 'critical'),
    (2, 'checkout', 1, 'critical'),
    (3, 'reports',  2, 'standard');
