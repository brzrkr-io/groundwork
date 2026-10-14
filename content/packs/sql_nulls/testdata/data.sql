INSERT INTO deploys (id, service, version, environment, status, started_at, finished_at, duration_s, deployed_by) VALUES
    (1,  'api',      'v1.4.0', 'staging',    'succeeded', '2026-03-02 09:12:00', '2026-03-02 09:13:35', 95,   'priya'),
    (2,  'api',      'v1.4.0', 'production', 'succeeded', '2026-03-02 10:05:00', '2026-03-02 10:07:20', 140,  'priya'),
    (3,  'checkout', 'v2.7.1', NULL,         'succeeded', '2026-03-02 11:30:00', '2026-03-02 11:33:30', 210,  NULL),
    (4,  'checkout', 'v2.7.1', 'production', 'failed',    '2026-03-02 13:02:00', '2026-03-02 13:12:10', 610,  'marco'),
    (5,  'billing',  'v0.9.3', 'production', 'succeeded', '2026-03-03 08:45:00', '2026-03-03 08:46:28', NULL, 'marco'),
    (6,  'search',   'v3.1.0', 'staging',    'failed',    '2026-03-03 09:20:00', '2026-03-03 09:20:45', 45,   NULL),
    (7,  'search',   'v3.1.1', NULL,         'succeeded', '2026-03-03 10:01:00', '2026-03-03 10:01:52', 52,   NULL),
    (8,  'billing',  'v1.0.0', 'staging',    'running',   '2026-03-06 16:58:00', NULL,                  NULL, 'marco'),
    (9,  'api',      'v1.5.0', 'production', 'running',   '2026-03-06 17:02:00', NULL,                  NULL, NULL),
    (10, 'checkout', 'v2.8.0', 'production', 'succeeded', '2026-03-06 10:00:00', '2026-03-06 10:07:36', 456,  'li');
