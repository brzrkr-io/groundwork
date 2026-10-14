-- Task 4: List each production service once
--
-- Show the name of every service that has had a production deploy,
-- each name once, A to Z.

SELECT DISTINCT service
FROM deploys
WHERE environment = 'production'
ORDER BY service;
