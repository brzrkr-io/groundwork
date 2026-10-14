## Task 1: List every service with its on-call engineer

Every service must appear, even with no match in `oncall`. That's the job of
the join that keeps all rows of the left table.

### Closer

`FROM services AS s LEFT JOIN oncall AS o ON o.service_id = s.id`. Select
`s.name AS service, o.engineer` and sort by `s.name`.

### Answer

```sql
SELECT s.name AS service, o.engineer
FROM services AS s
LEFT JOIN oncall AS o ON o.service_id = s.id
ORDER BY s.name;
```

## Task 2: Find the services nobody is on call for

Start from task 1's join. The services you want are exactly the rows where
the `oncall` side came back empty.

### Closer

Add `WHERE o.service_id IS NULL`. Not `= NULL`: that's never true.

### Answer

```sql
SELECT s.name AS service
FROM services AS s
LEFT JOIN oncall AS o ON o.service_id = s.id
WHERE o.service_id IS NULL
ORDER BY s.name;
```

## Task 3: Count deploys per service, including zero

Left-join `deploys` onto `services` so every service has at least one row,
then group per service. Count something that is `NULL` when there was no
deploy.

### Closer

`LEFT JOIN deploys AS d ON d.service_id = s.id`, `GROUP BY s.id`, and
`COUNT(d.id) AS deploys`.

### Answer

```sql
SELECT s.name AS service, COUNT(d.id) AS deploys
FROM services AS s
LEFT JOIN deploys AS d ON d.service_id = s.id
GROUP BY s.id
ORDER BY s.name;
```

## Task 4: Fix the production deploy count

The `WHERE` runs after the join. For `notifications`, `d.environment` is
`NULL`, which isn't `'production'`, so its only row is thrown away.

### Closer

Delete the `WHERE` line and add the environment check to the join itself:
`ON d.service_id = s.id AND ...`.

### Answer

```sql
SELECT s.name AS service, COUNT(d.id) AS deploys
FROM services AS s
LEFT JOIN deploys AS d
    ON d.service_id = s.id
    AND d.environment = 'production'
GROUP BY s.id
ORDER BY s.name;
```

In `ON`, the condition only decides which deploys match; every service still
gets its row.
