## Task 1: List the open incidents

Join `incidents` to `services` for the name, keep the incidents with no
`resolved_at`, then sort on two keys.

### Closer

`JOIN services AS s ON s.id = i.service_id`, `WHERE i.resolved_at IS NULL`,
`ORDER BY i.severity, i.opened_at`.

### Answer

```sql
SELECT i.id, s.name AS service, i.severity, i.opened_at
FROM incidents AS i
JOIN services AS s ON s.id = i.service_id
WHERE i.resolved_at IS NULL
ORDER BY i.severity, i.opened_at;
```

## Task 2: Report time to resolve per service

Every service must appear, so start from `services` and left-join the
incidents. Group per service, count a column from the incidents side, and let
`AVG` skip the open incidents' `NULL`s.

### Closer

`FROM services AS s LEFT JOIN incidents AS i ON i.service_id = s.id`,
`GROUP BY s.id`, with `COUNT(i.id)` and `ROUND(AVG(i.minutes_to_resolve), 1)`.

### Answer

```sql
SELECT
    s.name AS service,
    COUNT(i.id) AS incidents,
    ROUND(AVG(i.minutes_to_resolve), 1) AS avg_minutes
FROM services AS s
LEFT JOIN incidents AS i ON i.service_id = s.id
GROUP BY s.id
ORDER BY s.name;
```

## Task 3: Find the missing postmortems

"Has no postmortem" is the left-join-then-`IS NULL` pattern. The other two
conditions are an `IS NOT NULL` and an `IN`.

### Closer

Join `services` for the name, then
`LEFT JOIN postmortems AS p ON p.incident_id = i.id`. In the `WHERE`:
`i.resolved_at IS NOT NULL`, `i.severity IN (...)`, `p.incident_id IS NULL`.

### Answer

```sql
SELECT i.id, s.name AS service, i.severity
FROM incidents AS i
JOIN services AS s ON s.id = i.service_id
LEFT JOIN postmortems AS p ON p.incident_id = i.id
WHERE i.resolved_at IS NOT NULL
  AND i.severity IN ('sev1', 'sev2')
  AND p.incident_id IS NULL
ORDER BY i.id;
```

## Task 4: Rank the noisiest teams in March

Two inner joins get from each incident to its team. Filter the March rows in
`WHERE`, group per team, keep the groups with at least 2 in `HAVING`.

### Closer

`WHERE i.opened_at >= '2026-03-01' AND i.opened_at < '2026-04-01'`,
`GROUP BY t.id`, `HAVING COUNT(*) >= 2`, `ORDER BY incidents DESC, t.name`.

### Answer

```sql
SELECT t.name AS team, COUNT(*) AS incidents
FROM incidents AS i
JOIN services AS s ON s.id = i.service_id
JOIN teams AS t ON t.id = s.team_id
WHERE i.opened_at >= '2026-03-01'
  AND i.opened_at < '2026-04-01'
GROUP BY t.id
HAVING COUNT(*) >= 2
ORDER BY incidents DESC, t.name;
```

## Task 5: Close incident 7 with its postmortem

Two changes that belong together: wrap them in a transaction. The `UPDATE`
sets two columns at once, separated by a comma.

### Closer

`BEGIN;`, then `UPDATE incidents SET resolved_at = ..., minutes_to_resolve = ... WHERE id = 7;`,
then `INSERT INTO postmortems (incident_id, author, published_at) VALUES (...);`,
then `COMMIT;`.

### Answer

```sql
BEGIN;

UPDATE incidents
SET resolved_at = '2026-03-06 11:02:00',
    minutes_to_resolve = 57
WHERE id = 7;

INSERT INTO postmortems (incident_id, author, published_at)
VALUES (7, 'marco', '2026-03-09 10:00:00');

COMMIT;
```
