-- Task 2: Find the services nobody is on call for
--
-- Return the name (service) of every service with no on-call
-- engineer, A to Z.

SELECT s.name AS service
FROM services AS s
LEFT JOIN oncall AS o ON o.service_id = s.id
WHERE o.service_id IS NULL
ORDER BY s.name;
