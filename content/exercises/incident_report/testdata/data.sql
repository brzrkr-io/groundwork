INSERT INTO teams (id, name, slack_channel) VALUES
    (1, 'payments', '#payments-oncall'),
    (2, 'platform', '#platform-oncall'),
    (3, 'search',   '#search-team');

INSERT INTO services (id, name, team_id) VALUES
    (1, 'api',           2),
    (2, 'checkout',      1),
    (3, 'billing',       1),
    (4, 'search',        3),
    (5, 'notifications', 2);

INSERT INTO incidents (id, service_id, severity, title, opened_at, resolved_at, minutes_to_resolve) VALUES
    (1, 2, 'sev1', 'Card payments failing',      '2026-03-02 13:05:00', '2026-03-02 13:47:00', 42),
    (2, 1, 'sev2', 'Elevated 5xx on /login',     '2026-03-03 07:10:00', '2026-03-03 07:35:00', 25),
    (3, 4, 'sev3', 'Slow autocomplete',          '2026-03-03 09:30:00', '2026-03-03 11:00:00', 90),
    (4, 3, 'sev2', 'Invoices not sent',          '2026-03-04 17:00:00', '2026-03-04 18:10:00', 70),
    (5, 2, 'sev2', 'Cart timeouts',              '2026-03-05 12:00:00', '2026-03-05 12:18:00', 18),
    (6, 1, 'sev3', 'Noisy disk alert',           '2026-03-06 08:00:00', NULL,                  NULL),
    (7, 2, 'sev1', 'Checkout down after v2.8.0', '2026-03-06 10:05:00', NULL,                  NULL),
    (8, 3, 'sev3', 'Retry queue growing',        '2026-03-06 11:20:00', NULL,                  NULL),
    (9, 1, 'sev1', 'Database failover',          '2026-02-20 03:00:00', '2026-02-20 04:30:00', 90);

INSERT INTO postmortems (incident_id, author, published_at) VALUES
    (1, 'marco', '2026-03-04 15:00:00'),
    (4, 'li',    '2026-03-06 09:00:00'),
    (9, 'priya', '2026-02-23 10:00:00');
