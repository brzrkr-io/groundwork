-- Task 1: Fix the busy services query
--
-- Return each service with more than 3 deploys, and its number of
-- deploys (deploys), A to Z. This fails with an error.
--
-- TODO(1): Fix the query so it runs and keeps only the busy services.

SELECT service, COUNT(*) AS deploys
FROM deploys
WHERE COUNT(*) > 3
GROUP BY service
ORDER BY service;
