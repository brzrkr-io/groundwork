## Task 1: List every deploy

A query that wants every row needs only two parts: the columns to show and
the table they come from. List the columns in the order the task gives.

### Closer

`SELECT` takes a comma-separated list of column names; `FROM` takes the table
name, `deploys`. No `WHERE` needed.

### Answer

```sql
SELECT service, version, environment
FROM deploys;
```

## Task 2: Find the failed deploys

Start from a query that shows `service` and `version` for every row, then add
a condition that keeps only the failed ones.

### Closer

Add `WHERE status = ...` after the `FROM` line. `failed` is text, so it goes
in single quotes.

### Answer

```sql
SELECT service, version
FROM deploys
WHERE status = 'failed';
```

## Task 3: Find slow production deploys

Two things must both be true: the environment is production, and the
duration is over 300. Join two conditions with the word that means "both".

### Closer

`WHERE environment = 'production' AND ...`. `duration_s` is a number, so
compare it to `300` without quotes, using `>` (not `>=`).

### Answer

```sql
SELECT id, service, duration_s
FROM deploys
WHERE environment = 'production'
  AND duration_s > 300;
```

## Task 4: Find the payments team's problem deploys

You need two conditions joined with `AND`: the service is one of two names,
and the status is not `succeeded`. Statuses other than `failed` count too.

### Closer

`service IN ('checkout', 'billing')` matches either name. For the status, use
`!=` (or `<>`) with `'succeeded'`.

### Answer

```sql
SELECT id, service, status
FROM deploys
WHERE service IN ('checkout', 'billing')
  AND status != 'succeeded';
```

Checking "not succeeded" instead of "= 'failed'" also catches `rolled_back`.
