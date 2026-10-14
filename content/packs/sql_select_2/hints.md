## Task 1: Fix the staging filter

The query is fine except for one value. Compare the text in the `WHERE` line
with the way environments are written in `testdata/data.sql`.

### Closer

Text comparison with `=` is exact, including capital letters. The data says
`staging`, all lower case.

### Answer

```sql
SELECT service, version
FROM deploys
WHERE environment = 'staging';
```

## Task 2: Fix the production problems report

SQL reads the condition as `(production AND failed) OR rolled_back`, so any
rolled-back deploy passes, whatever its environment.

### Closer

Group the two statuses so the environment check applies to both: put
parentheses around the `OR`, or replace it with `status IN (...)`.

### Answer

```sql
SELECT id, service, status
FROM deploys
WHERE environment = 'production'
  AND status IN ('failed', 'rolled_back');
```

`environment = 'production' AND (status = 'failed' OR status = 'rolled_back')`
works too.

## Task 3: Fix the release candidate search

`%` only means "anything" for one operator. With `=`, the query looks for a
version that is literally `%-rc%`.

### Closer

Swap `=` for the pattern-matching operator, and keep the pattern as it is.

### Answer

```sql
SELECT service, version
FROM deploys
WHERE version LIKE '%-rc%';
```

## Task 4: Fix the date range

`BETWEEN '2026-03-03' AND '2026-03-04'` stops at the text `'2026-03-04'`,
and every time on March 4th is greater than that shorter string.

### Closer

Use a half-open range: `started_at >= ` the first day, `AND started_at < ` the
day after the last one.

### Answer

```sql
SELECT id, service, started_at
FROM deploys
WHERE started_at >= '2026-03-03'
  AND started_at < '2026-03-05';
```

The `<` on the next day includes every second of March 4th and nothing after.
