-- Task 4: Fix the slow services report
--
-- Return each service whose production deploys take more than 300
-- seconds on average, with that average in minutes rounded to one
-- decimal place (avg_minutes), A to Z. This averages only the slow
-- deploys, so billing looks almost twice as slow as it is.
--
-- TODO(4): Fix the query so it averages every production deploy,
-- then keeps the slow services.

SELECT service, ROUND(AVG(duration_s) / 60.0, 1) AS avg_minutes
FROM deploys
WHERE environment = 'production' AND duration_s > 300
GROUP BY service
ORDER BY service;
