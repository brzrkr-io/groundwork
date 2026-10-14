-- Task 1: List every service with its on-call engineer
--
-- Return every service's name (service) and its on-call engineer
-- (engineer), A to Z by service. A service with nobody on call
-- still appears, with NULL as its engineer.

SELECT s.name AS service, o.engineer
FROM services AS s
LEFT JOIN oncall AS o ON o.service_id = s.id
ORDER BY s.name;
