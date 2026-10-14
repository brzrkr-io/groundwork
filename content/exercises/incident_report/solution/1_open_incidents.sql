-- Task 1: List the open incidents
--
-- Return the id, service name (service), severity and opened_at of
-- every incident that is still open: worst severity first, and
-- oldest first within a severity.

SELECT i.id, s.name AS service, i.severity, i.opened_at
FROM incidents AS i
JOIN services AS s ON s.id = i.service_id
WHERE i.resolved_at IS NULL
ORDER BY i.severity, i.opened_at;
