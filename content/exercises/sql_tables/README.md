# CREATE TABLE: keys and constraints

A good table refuses bad data: a deploy for a service that doesn't exist, a
status nobody understands, a team with no name. Here you design the tables
for a deploy tracker, with the rules built in.

**New here:** `CREATE TABLE`, column types, `PRIMARY KEY`, `NOT NULL`, `UNIQUE`, `DEFAULT`, `CHECK`, `REFERENCES` (foreign keys), `ALTER TABLE … ADD COLUMN`.

## Learn

### CREATE TABLE

`CREATE TABLE` names the table, then lists its columns, each with a name and
a type, separated by commas:

```sql
CREATE TABLE hosts (
    id       INTEGER PRIMARY KEY,
    hostname TEXT NOT NULL UNIQUE,
    cpus     INTEGER NOT NULL DEFAULT 2
);
```

| Type | Holds |
|---|---|
| `INTEGER` | whole numbers |
| `REAL` | decimal numbers |
| `TEXT` | text (SQLite stores timestamps as text too) |

> **Why is it written like this?** SQL requires the parentheses and the
> commas between columns, and no comma after the last one (like a Go
> function's parameters, not like a Go composite literal). Lining the types
> up is a style choice. Table names in the plural (`hosts`) and columns in
> *snake_case* are the most common convention.

### Constraints: rules the database enforces

A *constraint* goes after the column's type. An `INSERT` or `UPDATE` that
breaks one fails with an error, and nothing is written.

| Constraint | Rule |
|---|---|
| `PRIMARY KEY` | identifies each row: unique and never empty. An `INTEGER PRIMARY KEY` is numbered for you (1, 2, 3, …) when an insert leaves it out. |
| `NOT NULL` | a value is required |
| `UNIQUE` | no two rows may share a value |
| `DEFAULT x` | the value used when an insert leaves the column out |
| `CHECK (condition)` | every row must satisfy the condition, e.g. `CHECK (cpus > 0)` |

A `CHECK` treats an unknown (`NULL`) result as passing (see *NULL: fix the
missing-value bugs*), so a column that must have a valid value needs
`NOT NULL` **and** the `CHECK`.

### Foreign keys

A *foreign key* says "this column holds the `id` of a row in that table":

```sql
CREATE TABLE disks (
    id      INTEGER PRIMARY KEY,
    host_id INTEGER NOT NULL REFERENCES hosts (id),
    size_gb INTEGER NOT NULL
);
```

Now a disk for host 99 is rejected unless host 99 exists, and a host that
still has disks can't be deleted. This is what keeps the joins from *JOINs:
services, teams and who to page* honest.

SQLite only enforces foreign keys after `PRAGMA foreign_keys = ON`, once per
connection; the tests do that for you. Postgres and MySQL always enforce
them.

### Rules over several columns

A constraint can also go on its own line after the columns, naming several
columns at once. Then the *combination* must be unique:

```sql
CREATE TABLE mounts (
    host_id INTEGER NOT NULL REFERENCES hosts (id),
    path    TEXT NOT NULL,
    UNIQUE (host_id, path)   -- one /data per host, but every host may have one
);
```

### Changing a table that has data: migrations

Once a table holds production data you can't just drop and recreate it. A
*migration* is a script that changes the schema in place, and it's checked
into Git and run by the deploy pipeline. The most common one adds a column:

```sql
ALTER TABLE hosts ADD COLUMN region TEXT;
```

Existing rows get the column's default, or `NULL` if there is none. That's
why a new `NOT NULL` column **must** have a `DEFAULT`: otherwise the rows
already there would break the rule, and SQLite refuses the change.

## Your task

You're setting up the tables for a new deploy tracker. Each file holds one
statement. The tests then try good and bad `INSERT`s, and tell you which rule
let bad data in. Write the columns in the order given.

### 1. Create the teams table

**Goal:** `1_teams_table.sql` creates a table `teams` with three columns:

| Column | Type | Rules |
|---|---|---|
| `id` | `INTEGER` | primary key |
| `name` | `TEXT` | required, unique |
| `slack_channel` | `TEXT` | required |

**Think it through:**
- Which constraint makes the database number the rows for you?
- "Required" and "unique" are two different constraints; one column can
  have both.

### 2. Create the services table

**Goal:** `2_services_table.sql` creates a table `services` with four
columns. A `teams` table already exists (from `testdata/teams.sql`).

| Column | Type | Rules |
|---|---|---|
| `id` | `INTEGER` | primary key |
| `name` | `TEXT` | required, unique |
| `team_id` | `INTEGER` | required, refers to `teams (id)` |
| `tier` | `TEXT` | required, either `'critical'` or `'standard'`, default `'standard'` |

**Think it through:**
- Which constraint points `team_id` at the `teams` table?
- `tier` has three rules. Which one rejects `'gold'`, which one rejects
  `NULL`?

### 3. Create the deploys table

**Goal:** `3_deploys_table.sql` creates a table `deploys` with five columns,
where no service has two deploys of the same version. `teams` and `services`
already exist.

| Column | Type | Rules |
|---|---|---|
| `id` | `INTEGER` | primary key |
| `service_id` | `INTEGER` | required, refers to `services (id)` |
| `version` | `TEXT` | required |
| `status` | `TEXT` | required, one of `'running'`, `'succeeded'`, `'failed'` |
| `started_at` | `TEXT` | required |

`v1.0.0` of `api` and `v1.0.0` of `checkout` are both fine; a second
`v1.0.0` of `api` is not.

**Think it through:**
- Should `version` alone be unique? What about the pair?
- Which operator lists the allowed statuses inside the `CHECK`?

### 4. Add a column with a migration

**Goal:** `4_add_slo_target.sql` adds a column `slo_target` to the existing
`services` table (from `testdata/services.sql`, which already has three
services): a `REAL`, required, with `99.9` for every service that doesn't
set its own, including the ones already there.

**Think it through:**
- Which statement changes an existing table without losing its rows?
- What value do existing rows get for a new column, and what does that mean
  for `NOT NULL`?

## Learn more

- [SQLite: CREATE TABLE](https://www.sqlite.org/lang_createtable.html)
- [SQLite: foreign key support](https://www.sqlite.org/foreignkeys.html)
- [SQLite: ALTER TABLE](https://www.sqlite.org/lang_altertable.html)
