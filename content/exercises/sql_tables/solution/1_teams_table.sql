-- Task 1: Create the teams table
--
-- Create a table named teams with three columns:
--   id             INTEGER, the primary key
--   name           TEXT, required and unique
--   slack_channel  TEXT, required

CREATE TABLE teams (
    id            INTEGER PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    slack_channel TEXT NOT NULL
);
