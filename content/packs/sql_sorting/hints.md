## Task 1: Sort the deploys by time

Select the three columns, then add the clause that sorts. Oldest first means
the smallest timestamp comes first.

### Closer

`ORDER BY started_at` goes after `FROM deploys`. Ascending is the default,
so you don't need `ASC`.

### Answer

```sql
SELECT id, service, started_at
FROM deploys
ORDER BY started_at;
```

## Task 2: Show the three latest production deploys

Three steps: keep production rows, sort newest first, keep three. Each step
is one clause, and SQL fixes their order.

### Closer

`WHERE environment = 'production'`, then `ORDER BY started_at DESC`, then
`LIMIT 3` last.

### Answer

```sql
SELECT id, service, version
FROM deploys
WHERE environment = 'production'
ORDER BY started_at DESC
LIMIT 3;
```

## Task 3: Sort by service, then slowest first

`ORDER BY` takes a list. The first column decides the order; the second only
matters between rows with the same service.

### Closer

`ORDER BY service, duration_s ...`. Each column has its own direction:
`service` ascending, `duration_s` descending.

### Answer

```sql
SELECT service, duration_s
FROM deploys
ORDER BY service, duration_s DESC;
```

## Task 4: List each production service once

Write the query that lists the service of every production deploy first,
then remove the repeats.

### Closer

`SELECT DISTINCT service` returns each value once. Add the production
`WHERE` and `ORDER BY service`.

### Answer

```sql
SELECT DISTINCT service
FROM deploys
WHERE environment = 'production'
ORDER BY service;
```
