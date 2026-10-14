-- Task 4: Fix the date range
--
-- Show the id, service and started_at of every deploy that started
-- on March 3rd or March 4th 2026. This misses every deploy from
-- March 4th.

SELECT id, service, started_at
FROM deploys
WHERE started_at >= '2026-03-03'
  AND started_at < '2026-03-05';
