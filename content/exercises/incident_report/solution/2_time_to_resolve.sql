-- Task 2: Report time to resolve per service
--
-- For every service, return its name (service), its number of
-- incidents (incidents) and the average minutes_to_resolve of its
-- resolved incidents, rounded to one decimal place (avg_minutes),
-- A to Z. A service with no incidents shows 0 and NULL.

SELECT
    s.name AS service,
    COUNT(i.id) AS incidents,
    ROUND(AVG(i.minutes_to_resolve), 1) AS avg_minutes
FROM services AS s
LEFT JOIN incidents AS i ON i.service_id = s.id
GROUP BY s.id
ORDER BY s.name;
