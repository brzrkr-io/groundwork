-- Task 3: Fix the version count
--
-- Return every service with the number of different versions it has
-- deployed (versions), A to Z. This counts a version twice when it
-- went to staging and then production.
--
-- TODO(3): Fix the query so each version counts once.

SELECT service, COUNT(version) AS versions
FROM deploys
GROUP BY service
ORDER BY service;
