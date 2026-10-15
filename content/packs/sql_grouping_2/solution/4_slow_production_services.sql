-- Task 4: Fix the slow services report
--
-- Return each service whose production deploys take more than 300
-- seconds on average, with that average in minutes rounded to one
-- decimal place (avg_minutes), A to Z. The starter query averaged
-- only the slow deploys, so billing looked almost twice as slow as it is.

SELECT service, ROUND(AVG(duration_s) / 60.0, 1) AS avg_minutes
FROM deploys
WHERE environment = 'production'
GROUP BY service
HAVING AVG(duration_s) > 300
ORDER BY service;
