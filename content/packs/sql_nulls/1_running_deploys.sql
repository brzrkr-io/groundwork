-- Task 1: Fix the running deploys query
--
-- Return the id, service and started_at of every deploy that is
-- still running (it has no finished_at), by id. This returns no rows.
--
-- TODO(1): Fix the test for a missing finished_at.

SELECT id, service, started_at
FROM deploys
WHERE finished_at = NULL
ORDER BY id;
