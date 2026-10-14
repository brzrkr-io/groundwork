-- Task 2: Fix the failure counts
--
-- Return every service with its number of deploys (deploys) and
-- its number of failed deploys (failures), A to Z. This drops api
-- and counts only failures in both columns.
--
-- TODO(2): Fix the query so deploys counts every deploy and failures only the failed ones.

SELECT service, COUNT(*) AS deploys, COUNT(*) AS failures
FROM deploys
WHERE status = 'failed'
GROUP BY service
ORDER BY service;
