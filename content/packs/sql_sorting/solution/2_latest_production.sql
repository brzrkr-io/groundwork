-- Task 2: Show the three latest production deploys
--
-- Show the id, service and version of the three most recent
-- production deploys, newest first.

SELECT id, service, version
FROM deploys
WHERE environment = 'production'
ORDER BY started_at DESC
LIMIT 3;
