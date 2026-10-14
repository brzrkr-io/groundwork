-- Task 1: Fix the staging filter
--
-- Show the service and version of every staging deploy.
-- This returns no rows at all.

SELECT service, version
FROM deploys
WHERE environment = 'staging';
