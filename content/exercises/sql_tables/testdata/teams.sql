-- A finished teams table, so the services and deploys tasks don't depend on
-- your answer to task 1.
CREATE TABLE teams (
    id            INTEGER PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    slack_channel TEXT NOT NULL
);

INSERT INTO teams (id, name, slack_channel) VALUES
    (1, 'payments', '#payments-oncall'),
    (2, 'platform', '#platform-oncall');
