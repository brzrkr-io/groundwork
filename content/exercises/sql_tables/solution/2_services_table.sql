-- Task 2: Create the services table
--
-- Create a table named services with four columns:
--   id       INTEGER, the primary key
--   name     TEXT, required and unique
--   team_id  INTEGER, required, a foreign key to teams (id)
--   tier     TEXT, required, 'critical' or 'standard', default 'standard'

CREATE TABLE services (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    team_id INTEGER NOT NULL REFERENCES teams (id),
    tier    TEXT NOT NULL DEFAULT 'standard' CHECK (tier IN ('critical', 'standard'))
);
