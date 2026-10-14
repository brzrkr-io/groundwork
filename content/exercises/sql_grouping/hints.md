## Task 1: Summarize every deploy

No grouping here: three aggregates over the whole table give exactly one
row. Pick the right function for each value, then name each column.

### Closer

`COUNT(*)` counts rows, `MAX(duration_s)` finds the largest, and `MIN` works
on the timestamps too. Name each with `AS`: `COUNT(*) AS deploys`.

### Answer

```sql
SELECT
    COUNT(*) AS deploys,
    MAX(duration_s) AS slowest_s,
    MIN(started_at) AS first_started
FROM deploys;
```

## Task 2: Count failures per service

First keep only the failed rows, then make one group per service and count
each group.

### Closer

`WHERE status = 'failed'` comes before `GROUP BY service`. Select `service`
and `COUNT(*) AS failures`, then `ORDER BY service`.

### Answer

```sql
SELECT service, COUNT(*) AS failures
FROM deploys
WHERE status = 'failed'
GROUP BY service
ORDER BY service;
```

## Task 3: Average the duration per environment

Group by environment; each group then gets its own count and average. Round
the average, not the individual durations.

### Closer

`COUNT(*) AS deploys` and `ROUND(AVG(duration_s), 1) AS avg_duration_s`, with
`GROUP BY environment` and `ORDER BY environment`.

### Answer

```sql
SELECT
    environment,
    COUNT(*) AS deploys,
    ROUND(AVG(duration_s), 1) AS avg_duration_s
FROM deploys
GROUP BY environment
ORDER BY environment;
```

## Task 4: Count deploys per day

Every timestamp is different, so grouping by `started_at` gives one group
per deploy. Group by the date part instead.

### Closer

`date(started_at) AS day` makes the day column. Then `GROUP BY day` and
`ORDER BY day DESC`.

### Answer

```sql
SELECT date(started_at) AS day, COUNT(*) AS deploys
FROM deploys
GROUP BY day
ORDER BY day DESC;
```
