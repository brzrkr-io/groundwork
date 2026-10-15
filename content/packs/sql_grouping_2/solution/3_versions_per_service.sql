-- Task 3: Fix the version count
--
-- Return every service with the number of different versions it has
-- deployed (versions), A to Z. The starter query counted a version
-- twice when it went to staging and then production.

SELECT service, COUNT(DISTINCT version) AS versions
FROM deploys
GROUP BY service
ORDER BY service;
