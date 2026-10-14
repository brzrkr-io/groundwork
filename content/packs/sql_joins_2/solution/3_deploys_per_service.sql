-- Task 3: Count deploys per service, including zero
--
-- Return every service's name (service) and its number of deploys
-- (deploys), A to Z. A service that never deployed shows 0.

SELECT s.name AS service, COUNT(d.id) AS deploys
FROM services AS s
LEFT JOIN deploys AS d ON d.service_id = s.id
GROUP BY s.id
ORDER BY s.name;
