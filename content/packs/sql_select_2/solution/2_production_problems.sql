-- Task 2: Fix the production problems report
--
-- Show the id, service and status of every production deploy that
-- failed or was rolled back. This also returns a staging deploy.

SELECT id, service, status
FROM deploys
WHERE environment = 'production'
  AND status IN ('failed', 'rolled_back');
