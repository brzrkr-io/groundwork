## Task 1: Create the teams table

Write `CREATE TABLE teams (...)` with one line per column: name, type, then
its constraints. Read the table in the README one row at a time.

### Closer

`id INTEGER PRIMARY KEY`, then `name TEXT` followed by both `NOT NULL` and
`UNIQUE`, then `slack_channel TEXT NOT NULL`. Commas between columns, none
after the last.

### Answer

```sql
CREATE TABLE teams (
    id            INTEGER PRIMARY KEY,
    name          TEXT NOT NULL UNIQUE,
    slack_channel TEXT NOT NULL
);
```

## Task 2: Create the services table

Same shape as `teams`, plus two new ideas: a column that points at another
table, and a column with a default and an allowed list of values.

### Closer

`team_id INTEGER NOT NULL REFERENCES teams (id)`. For the tier:
`tier TEXT NOT NULL DEFAULT 'standard' CHECK (tier IN (...))`.

### Answer

```sql
CREATE TABLE services (
    id      INTEGER PRIMARY KEY,
    name    TEXT NOT NULL UNIQUE,
    team_id INTEGER NOT NULL REFERENCES teams (id),
    tier    TEXT NOT NULL DEFAULT 'standard' CHECK (tier IN ('critical', 'standard'))
);
```

## Task 3: Create the deploys table

Each column follows the same pattern as before. The "no two deploys of the
same version" rule is about two columns together, so it can't go on one
column.

### Closer

After the last column, add a comma and a table constraint:
`UNIQUE (service_id, version)`. The status rule is
`CHECK (status IN ('running', 'succeeded', 'failed'))` plus `NOT NULL`.

### Answer

```sql
CREATE TABLE deploys (
    id         INTEGER PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services (id),
    version    TEXT NOT NULL,
    status     TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
    started_at TEXT NOT NULL,
    UNIQUE (service_id, version)
);
```

## Task 4: Add a column with a migration

Recreating the table would lose its rows. Change it in place, and give the
new column a default so the existing rows have a value.

### Closer

`ALTER TABLE services ADD COLUMN slo_target REAL ...` followed by the
`NOT NULL` and `DEFAULT 99.9` constraints.

### Answer

```sql
ALTER TABLE services ADD COLUMN slo_target REAL NOT NULL DEFAULT 99.9;
```

Without the `DEFAULT`, the three existing rows would break `NOT NULL`, and
SQLite refuses the change.
