-- Task 4: Fix the production deploy count
--
-- Return every service's name (service) and its number of production
-- deploys (deploys), A to Z. A service with none shows 0. This loses
-- the services with no production deploys.
--
-- TODO(4): Fix the query so every service appears.

SELECT s.name AS service, COUNT(d.id) AS deploys
FROM services AS s
LEFT JOIN deploys AS d ON d.service_id = s.id
WHERE d.environment = 'production'
GROUP BY s.id
ORDER BY s.name;
