## Task 1: Record a new deploy

An `INSERT` names the table, lists the columns you're filling, then gives the
values in the same order. Leave `id` out of both lists.

### Closer

`INSERT INTO deploys (service_id, version, environment, status, started_at)`
then `VALUES (...)`. The number `4` needs no quotes; the text values do.

### Answer

```sql
INSERT INTO deploys (service_id, version, environment, status, started_at)
VALUES (4, 'v3.2.0', 'staging', 'running', '2026-03-07 09:00:00');
```

## Task 2: Mark one deploy as rolled back

An `UPDATE` needs three parts: the table, the new value, and which row. The
deploy's `id` picks exactly one row.

### Closer

`UPDATE deploys SET status = ...` and then `WHERE id = 14`.

### Answer

```sql
UPDATE deploys
SET status = 'rolled_back'
WHERE id = 14;
```

## Task 3: Delete last year's staging deploys

Write the `WHERE` first: staging **and** started before the first moment of
2026. Timestamps compare as text, so compare with `'2026-01-01'`.

### Closer

`DELETE FROM deploys WHERE environment = 'staging' AND started_at < ...`.
Read it as `SELECT id FROM deploys WHERE ...` against `testdata/data.sql`:
it should pick only 15 and 16.

### Answer

```sql
DELETE FROM deploys
WHERE environment = 'staging'
  AND started_at < '2026-01-01';
```

## Task 4: Release checkout in one transaction

You already know both statements: an `INSERT` like task 1 and an `UPDATE`
like task 2. The new part is wrapping them so they succeed or fail together.

### Closer

`BEGIN;` on the first line, then the `INSERT`, then
`UPDATE services SET current_version = ... WHERE id = 2;`, then `COMMIT;`.

### Answer

```sql
BEGIN;

INSERT INTO deploys (service_id, version, environment, status, started_at)
VALUES (2, 'v2.8.1', 'production', 'succeeded', '2026-03-07 10:00:00');

UPDATE services
SET current_version = 'v2.8.1'
WHERE id = 2;

COMMIT;
```

If the `UPDATE` fails, the `COMMIT` never runs and the `ROLLBACK` that
follows the error undoes the `INSERT` too.
