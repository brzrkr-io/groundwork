-- Task 2: Fix the production problems report
--
-- Show the id, service and status of every production deploy that
-- failed or was rolled back. This also returns a staging deploy.
--
-- TODO(2): Fix the query so only production deploys come back.

SELECT id, service, status
FROM deploys
WHERE environment = 'production' AND status = 'failed' OR status = 'rolled_back';
