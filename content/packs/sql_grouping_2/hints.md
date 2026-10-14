## Task 1: Fix the busy services query

`WHERE` runs on single rows, before any group exists, so it can't count
anything yet. The count is only known once the groups are made.

### Closer

Remove the `WHERE` line and put `HAVING COUNT(*) > 3` right after
`GROUP BY service`, before `ORDER BY`.

### Answer

```sql
SELECT service, COUNT(*) AS deploys
FROM deploys
GROUP BY service
HAVING COUNT(*) > 3
ORDER BY service;
```

## Task 2: Fix the failure counts

The `WHERE` keeps only failed rows, so every column only sees failures, and a
service with no failures disappears. Count all rows, and make `failures`
pick out the failed ones itself.

### Closer

Delete the `WHERE`. For `failures`, sum a 1 for each failed row and a 0 for
the rest: `SUM(CASE WHEN status = 'failed' THEN ... ELSE ... END)`.

### Answer

```sql
SELECT
    service,
    COUNT(*) AS deploys,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failures
FROM deploys
GROUP BY service
ORDER BY service;
```

## Task 3: Fix the version count

`COUNT(version)` counts rows that have a version: four api deploys give 4.
You want the number of different values.

### Closer

Put `DISTINCT` inside the brackets: `COUNT(DISTINCT ...)`.

### Answer

```sql
SELECT service, COUNT(DISTINCT version) AS versions
FROM deploys
GROUP BY service
ORDER BY service;
```

## Task 4: Fix the slow services report

`duration_s > 300` in `WHERE` throws away the fast deploys before averaging.
Average every production deploy, then keep the groups whose average is over
300.

### Closer

Keep `WHERE environment = 'production'`. Move the 300 check to
`HAVING AVG(duration_s) > 300`, after `GROUP BY`.

### Answer

```sql
SELECT service, ROUND(AVG(duration_s) / 60.0, 1) AS avg_minutes
FROM deploys
WHERE environment = 'production'
GROUP BY service
HAVING AVG(duration_s) > 300
ORDER BY service;
```
