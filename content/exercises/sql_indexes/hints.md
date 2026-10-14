## Task 1: Index the service column

An index needs a name, the table it belongs to, and the column(s) it sorts.
The task gives you all three.

### Closer

`CREATE INDEX <name> ON <table> (<column>);`

### Answer

```sql
CREATE INDEX deploys_by_service ON deploys (service);
```

## Task 2: Index the latest deploys of a service

The query filters with `= 'api'` on one column and sorts by another. Put the
`=` column first: then all of api's rows sit together, already in time order.

### Closer

`CREATE INDEX deploys_by_service_time ON deploys (service, ...)`. Reading the
index backwards gives the `DESC` order for free.

### Answer

```sql
CREATE INDEX deploys_by_service_time ON deploys (service, started_at);
```

`(started_at, service)` would sort by time first, scattering api's rows, so
SQLite would have to scan.

## Task 3: Stop duplicate deploy records

Same statement as task 1, with one extra word that makes the database refuse
a second row with the same values, and three columns.

### Closer

`CREATE UNIQUE INDEX deploys_one_per_release ON deploys (...)`, listing
`service`, `version`, `environment` in that order.

### Answer

```sql
CREATE UNIQUE INDEX deploys_one_per_release ON deploys (service, version, environment);
```

## Task 4: Rewrite the query so it can use the index

The index holds `started_at` values, not `date(started_at)` values. Describe
the day as a range of bare `started_at` values instead.

### Closer

Replace the `WHERE` with `started_at >= '2026-03-04' AND started_at < ...`,
the first moment of the next day. Keep the `ORDER BY`.

### Answer

```sql
SELECT id, service
FROM deploys
WHERE started_at >= '2026-03-04'
  AND started_at < '2026-03-05'
ORDER BY started_at;
```

Now SQLite jumps to the first March 4th entry and stops at March 5th.
