-- Task 4: Fix the duration report
--
-- For each service, return its number of deploys (deploys) and the
-- average of its recorded durations rounded to one decimal place
-- (avg_duration_s), A to Z. Deploys without a duration count as
-- deploys but must not pull the average down.

SELECT
    service,
    COUNT(*) AS deploys,
    ROUND(AVG(duration_s), 1) AS avg_duration_s
FROM deploys
GROUP BY service
ORDER BY service;
