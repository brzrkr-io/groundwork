-- Task 1: Fix the staging filter
--
-- Show the service and version of every staging deploy.
-- This returns no rows at all.
--
-- TODO(1): Fix the query so it finds the staging deploys.

SELECT service, version
FROM deploys
WHERE environment = 'Staging';
