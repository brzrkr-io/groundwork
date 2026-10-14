-- Task 4: Release checkout in one transaction
--
-- In one transaction: record a succeeded production deploy of
-- checkout (service_id 2), version v2.8.1, started at
-- 2026-03-07 10:00:00; then set checkout's current_version to v2.8.1.

BEGIN;

INSERT INTO deploys (service_id, version, environment, status, started_at)
VALUES (2, 'v2.8.1', 'production', 'succeeded', '2026-03-07 10:00:00');

UPDATE services
SET current_version = 'v2.8.1'
WHERE id = 2;

COMMIT;
