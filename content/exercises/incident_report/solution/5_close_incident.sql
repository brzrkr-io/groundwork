-- Task 5: Close incident 7 with its postmortem
--
-- In one transaction: mark incident 7 resolved at 2026-03-06 11:02:00
-- after 57 minutes, then add its postmortem by marco, published at
-- 2026-03-09 10:00:00.

BEGIN;

UPDATE incidents
SET resolved_at = '2026-03-06 11:02:00',
    minutes_to_resolve = 57
WHERE id = 7;

INSERT INTO postmortems (incident_id, author, published_at)
VALUES (7, 'marco', '2026-03-09 10:00:00');

COMMIT;
