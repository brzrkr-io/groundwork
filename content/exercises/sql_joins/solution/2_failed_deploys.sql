-- Task 2: Name the service of each failed deploy
--
-- Return the id, service name (service) and version of every failed
-- deploy, by deploy id.

SELECT d.id, s.name AS service, d.version
FROM deploys AS d
JOIN services AS s ON s.id = d.service_id
WHERE d.status = 'failed'
ORDER BY d.id;
