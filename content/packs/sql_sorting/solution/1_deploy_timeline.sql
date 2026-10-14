-- Task 1: Sort the deploys by time
--
-- Show the id, service and started_at of every deploy, oldest first.

SELECT id, service, started_at
FROM deploys
ORDER BY started_at;
