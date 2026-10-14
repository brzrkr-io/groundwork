-- Task 2: Count failures per service
--
-- For each service with at least one failed deploy, return the
-- service and its number of failed deploys (failures), A to Z.

SELECT service, COUNT(*) AS failures
FROM deploys
WHERE status = 'failed'
GROUP BY service
ORDER BY service;
