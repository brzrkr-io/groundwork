-- Task 4: Count deploys per day
--
-- For each day with deploys, return the day as YYYY-MM-DD (day) and
-- the number of deploys that started that day (deploys), newest day
-- first.

SELECT date(started_at) AS day, COUNT(*) AS deploys
FROM deploys
GROUP BY day
ORDER BY day DESC;
