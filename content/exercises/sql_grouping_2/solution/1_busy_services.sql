-- Task 1: Fix the busy services query
--
-- Return each service with more than 3 deploys, and its number of
-- deploys (deploys), A to Z. This fails with an error.

SELECT service, COUNT(*) AS deploys
FROM deploys
GROUP BY service
HAVING COUNT(*) > 3
ORDER BY service;
