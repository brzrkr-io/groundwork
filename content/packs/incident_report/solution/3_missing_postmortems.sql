-- Task 3: Find the missing postmortems
--
-- Return the id, service name (service) and severity of every
-- resolved sev1 or sev2 incident that has no postmortem, by id.

SELECT i.id, s.name AS service, i.severity
FROM incidents AS i
JOIN services AS s ON s.id = i.service_id
LEFT JOIN postmortems AS p ON p.incident_id = i.id
WHERE i.resolved_at IS NOT NULL
  AND i.severity IN ('sev1', 'sev2')
  AND p.incident_id IS NULL
ORDER BY i.id;
