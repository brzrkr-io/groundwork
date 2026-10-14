-- Task 1: Fix the running deploys query
--
-- Return the id, service and started_at of every deploy that is
-- still running (it has no finished_at), by id. This returns no rows.

SELECT id, service, started_at
FROM deploys
WHERE finished_at IS NULL
ORDER BY id;
