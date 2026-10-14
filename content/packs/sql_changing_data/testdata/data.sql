INSERT INTO services (id, name, current_version) VALUES
    (1, 'api',      'v1.4.1-hotfix'),
    (2, 'checkout', 'v2.7.2'),
    (3, 'billing',  'v0.9.3'),
    (4, 'search',   'v3.1.1');

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
    (14, 2, 'v2.8.0',        'production', 'failed',      '2026-03-06 10:00:00'),
    (15, 1, 'v1.3.0',        'staging',    'succeeded',   '2025-12-10 10:00:00'),
    (16, 2, 'v2.6.0',        'staging',    'failed',      '2025-12-15 14:00:00'),
    (17, 1, 'v1.3.0',        'production', 'succeeded',   '2025-12-11 09:00:00'),
    (18, 2, 'v2.6.1',        'production', 'succeeded',   '2025-12-16 11:00:00');
