-- Task 2: Fix the failure counts
--
-- Return every service with its number of deploys (deploys) and
-- its number of failed deploys (failures), A to Z. This drops api
-- and counts only failures in both columns.

SELECT
    service,
    COUNT(*) AS deploys,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failures
FROM deploys
GROUP BY service
ORDER BY service;
