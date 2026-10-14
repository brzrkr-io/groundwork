-- Task 3: Average the duration per environment
--
-- For each environment, return the environment, its number of
-- deploys (deploys) and its average duration_s rounded to one
-- decimal place (avg_duration_s), A to Z by environment.

SELECT
    environment,
    COUNT(*) AS deploys,
    ROUND(AVG(duration_s), 1) AS avg_duration_s
FROM deploys
GROUP BY environment
ORDER BY environment;
