INSERT INTO teams (id, name, slack_channel) VALUES
    (1, 'payments', '#payments-oncall'),
    (2, 'platform', '#platform-oncall'),
    (3, 'search',   '#search-team'),
    (4, 'data',     '#data-eng');

INSERT INTO services (id, name, team_id, tier) VALUES
    (1, 'api',           2,    'critical'),
    (2, 'checkout',      1,    'critical'),
    (3, 'billing',       1,    'critical'),
    (4, 'search',        3,    'standard'),
    (5, 'notifications', 2,    'standard'),
    (6, 'reports',       NULL, 'standard');

INSERT INTO oncall (service_id, engineer) VALUES
    (1, 'priya'),
    (2, 'marco'),
    (3, 'marco'),
    (5, 'li');

INSERT INTO deploys (id, service_id, version, environment, status, started_at) VALUES
    (1,  1, 'v1.4.0',        'staging',    'succeeded',   '2026-03-02 09:12:00'),
    (2,  1, 'v1.4.0',        'production', 'succeeded',   '2026-03-02 10:05:00'),
    (3,  2, 'v2.7.1',        'staging',    'succeeded',   '2026-03-02 11:30:00'),
    (4,  2, 'v2.7.1',        'production', 'failed',      '2026-03-02 13:02:00'),
    (5,  3, 'v0.9.3',        'production', 'succeeded',   '2026-03-03 08:45:00'),
    (6,  4, 'v3.1.0',        'staging',    'failed',      '2026-03-03 09:20:00'),
    (7,  4, 'v3.1.1',        'staging',    'succeeded',   '2026-03-03 10:01:00'),
    (8,  2, 'v2.7.2',        'production', 'succeeded',   '2026-03-03 14:40:00'),
    (9,  1, 'v1.5.0-rc.1',   'staging',    'succeeded',   '2026-03-04 09:00:00'),
    (10, 3, 'v1.0.0',        'staging',    'failed',      '2026-03-04 11:15:00'),
    (11, 3, 'v1.0.0',        'production', 'rolled_back', '2026-03-04 16:50:00'),
    (12, 4, 'v3.1.1',        'production', 'succeeded',   '2026-03-05 07:30:00'),
    (13, 1, 'v1.4.1-hotfix', 'production', 'succeeded',   '2026-03-05 22:10:00'),
    (14, 2, 'v2.8.0',        'production', 'failed',      '2026-03-06 10:00:00');
