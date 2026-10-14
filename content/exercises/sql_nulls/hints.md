## Task 1: Fix the running deploys query

Comparing anything with `NULL` using `=` gives "unknown", and `WHERE` drops
unknown rows. So this filter drops every row, including the running ones.

### Closer

SQL has a special operator for "is missing": `IS NULL`.

### Answer

```sql
SELECT id, service, started_at
FROM deploys
WHERE finished_at IS NULL
ORDER BY id;
```

## Task 2: Fix the non-production filter

For a deploy with no environment, `environment != 'production'` is unknown,
not true, so the row is dropped. You have to ask for those rows explicitly.

### Closer

Combine two conditions with `OR`: one for a missing environment
(`IS NULL`), one for "not production".

### Answer

```sql
SELECT id, service, environment
FROM deploys
WHERE environment IS NULL
   OR environment != 'production'
ORDER BY id;
```

In SQLite, `WHERE environment IS NOT 'production'` does the same in one test.

## Task 3: Fix the deployer labels

Joining text with `NULL` gives `NULL`. Replace the missing name with
`'automation'` **before** joining it to `'deployed by '`.

### Closer

`COALESCE(deployed_by, 'automation')` gives the name, or `'automation'` when
there is none. Use it in place of `deployed_by` after `||`.

### Answer

```sql
SELECT id, 'deployed by ' || COALESCE(deployed_by, 'automation') AS label
FROM deploys
ORDER BY id;
```

## Task 4: Fix the duration report

`COUNT(duration_s)` skips deploys without a duration, but they're still
deploys. And `SUM(...) / COUNT(*)` treats a missing duration as 0 seconds.

### Closer

Use `COUNT(*)` for `deploys`. For the average, `AVG(duration_s)` already
ignores `NULL`s; round it with `ROUND(..., 1)`.

### Answer

```sql
SELECT
    service,
    COUNT(*) AS deploys,
    ROUND(AVG(duration_s), 1) AS avg_duration_s
FROM deploys
GROUP BY service
ORDER BY service;
```
