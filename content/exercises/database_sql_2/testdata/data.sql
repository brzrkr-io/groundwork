INSERT INTO services (name, current_version) VALUES
    ('api',      'v1.4.1-hotfix'),
    ('checkout', 'v2.7.2'),
    ('billing',  'v0.9.3'),
    ('search',   'v3.1.1');

INSERT INTO deploys (id, service, version, environment, status, started_at, finished_at) VALUES
    (1,  'api',      'v1.4.0',        'staging',    'succeeded',   '2026-03-02 09:12:00', '2026-03-02 09:13:35'),
    (2,  'api',      'v1.4.0',        'production', 'succeeded',   '2026-03-02 10:05:00', '2026-03-02 10:07:20'),
    (3,  'checkout', 'v2.7.1',        'staging',    'succeeded',   '2026-03-02 11:30:00', '2026-03-02 11:33:30'),
    (4,  'checkout', 'v2.7.1',        'production', 'failed',      '2026-03-02 13:02:00', '2026-03-02 13:12:10'),
    (5,  'billing',  'v0.9.3',        'production', 'succeeded',   '2026-03-03 08:45:00', '2026-03-03 08:46:28'),
    (6,  'search',   'v3.1.0',        'staging',    'failed',      '2026-03-03 09:20:00', '2026-03-03 09:20:45'),
    (7,  'search',   'v3.1.1',        'staging',    'succeeded',   '2026-03-03 10:01:00', '2026-03-03 10:01:52'),
    (8,  'checkout', 'v2.7.2',        'production', 'succeeded',   '2026-03-03 14:40:00', '2026-03-03 14:46:35'),
    (9,  'api',      'v1.5.0-rc.1',   'staging',    'succeeded',   '2026-03-04 09:00:00', '2026-03-04 09:01:41'),
    (10, 'billing',  'v1.0.0',        'staging',    'failed',      '2026-03-04 11:15:00', '2026-03-04 11:20:02'),
    (11, 'billing',  'v1.0.0',        'production', 'rolled_back', '2026-03-04 16:50:00', '2026-03-04 17:02:00'),
    (12, 'search',   'v3.1.1',        'production', 'succeeded',   '2026-03-05 07:30:00', '2026-03-05 07:31:01'),
    (13, 'api',      'v1.4.1-hotfix', 'production', 'succeeded',   '2026-03-05 22:10:00', '2026-03-05 22:11:17'),
    (14, 'checkout', 'v2.8.0',        'production', 'failed',      '2026-03-06 10:00:00', '2026-03-06 10:07:35'),
    (15, 'search',   'v3.2.0',        'staging',    'running',     '2026-03-07 09:00:00', NULL);
