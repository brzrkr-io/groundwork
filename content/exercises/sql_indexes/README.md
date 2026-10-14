# Indexes: make the deploy lookups fast

With 14 rows every query is instant. With 40 million, "show me api's last
five deploys" can take seconds, and the dashboard that runs it every few
seconds takes the database down. An *index* turns that into a handful of
reads. Here you add indexes and check with `EXPLAIN QUERY PLAN` that SQLite
really uses them.

**New here:** `CREATE INDEX`, multi-column indexes, `CREATE UNIQUE INDEX`, `EXPLAIN QUERY PLAN`, `SCAN` vs `SEARCH`.

## Learn

### Without an index: read every row

To answer `WHERE service = 'api'`, a table with no index must look at every
row. That's a *full table scan*, O(n): twice the rows, twice the time (the
*Algorithms and Big-O* guide explains the notation).

### An index is a sorted copy of some columns

```sql
CREATE INDEX deploys_by_service ON deploys (service);
```

This builds a separate structure (a *B-tree*) holding every `service` value
in sorted order, each pointing at its row. Finding `'api'` is then like
finding a word in a dictionary: a few jumps, O(log n), whether the table has
a thousand rows or a billion.

The cost: each index takes disk space, and every `INSERT`, `UPDATE` and
`DELETE` must update it too. So index the columns your queries filter, join
and sort on, not everything.

> **Why is it written like this?** SQL requires the index name, the table
> and the columns in parentheses. The name is your choice; a convention
> like `table_by_columns` tells the next engineer what it's for when it shows
> up in a query plan.

### EXPLAIN QUERY PLAN: ask before you guess

Put `EXPLAIN QUERY PLAN` in front of a query and SQLite tells you how it
*would* run it, without running it:

```sql
EXPLAIN QUERY PLAN SELECT * FROM deploys WHERE environment = 'staging';
-- SCAN deploys
```

| Plan step | Means |
|---|---|
| `SCAN deploys` | reads every row: O(n) |
| `SEARCH deploys USING INDEX x (col=?)` | jumps straight to the matching rows through index `x` |
| `SCAN deploys USING INDEX x` | walks the whole index in order: still every row |
| `USE TEMP B-TREE FOR ORDER BY` | collects the rows, then sorts them as an extra step |

The tests print the plan for you whenever an index isn't used.

### Multi-column indexes: order matters

An index on `(service, started_at)` is sorted by `service`, and within each
service by `started_at`, like a phone book sorted by last name, then first
name.

```text
service  | started_at
api      | 2026-03-02 09:12:00
api      | 2026-03-02 10:05:00
api      | 2026-03-05 22:10:00
billing  | 2026-03-03 08:45:00
...
```

It helps queries that filter on the **leftmost** column(s): `WHERE service =
'api'` jumps to the api block, and the rows there are already in time order,
so `ORDER BY started_at` needs no extra sort step. It doesn't help
`WHERE started_at > ...` alone, just as a phone book doesn't help you find
everyone called "Sam".

### Unique indexes

`CREATE UNIQUE INDEX` also refuses duplicates, like the `UNIQUE` constraint
in *CREATE TABLE: keys and constraints*. It's how you add that rule to a
table that already exists.

### Hiding a column from its index

An index stores the column's values, not the results of functions on them.
`WHERE lower(service) = 'api'` or `WHERE substr(started_at, 1, 7) =
'2026-03'` can't search the index, because the index has no idea what
`lower(...)` gives. Compare the bare column with a value or a range instead.

## Your task

The deploy dashboard got slow. Make its lookups fast, and prove it with the
plan. The `deploys` table (`testdata/schema.sql`) already has one index,
`deploys_by_start`, on `started_at`.

### 1. Index the service column

**Goal:** `1_index_service.sql` creates an index named `deploys_by_service`
on the `service` column of `deploys`.

Afterwards the plan for `SELECT * FROM deploys WHERE service = 'api'` must
contain `SEARCH deploys USING INDEX deploys_by_service`.

**Think it through:**
- Three names in the statement: the index's, the table's, and the column's.

### 2. Index the latest deploys of a service

**Goal:** `2_index_latest.sql` creates an index named
`deploys_by_service_time` that lets this query **search** (not scan) **and**
skip the separate sort:

```sql
SELECT id, version FROM deploys
WHERE service = 'api'
ORDER BY started_at DESC
LIMIT 5;
```

The plan must contain `SEARCH deploys USING INDEX deploys_by_service_time`
and no `USE TEMP B-TREE`.

**Think it through:**
- Which column does the query filter on with `=`, and which does it sort by?
- Which of the two must come first in the index?

### 3. Stop duplicate deploy records

**Goal:** `3_unique_release.sql` creates a **unique** index named
`deploys_one_per_release` on `service`, `version` and `environment`, in that
order, so the same version of a service can't be recorded twice for one
environment.

Afterwards a second `production` row for `api` `v1.4.0` is rejected, while a
new version still goes in.

**Think it through:**
- Which word turns a plain index into one that rejects duplicates?

### 4. Rewrite the query so it can use the index

**Goal:** `4_one_day.sql` returns the `id` and `service` of every deploy
that started on **2026-03-04**, earliest first, and its plan must contain
`SEARCH deploys USING INDEX deploys_by_start`.

```text
9  | api
10 | billing
11 | billing
```

The starter's query returns the right rows, but `date(started_at)` hides the
column from the index, so the plan is `SCAN deploys USING INDEX
deploys_by_start`: it walks every entry.

**Think it through:**
- How do you say "on March 4th" with the bare `started_at` column, as a
  range (see *SELECT and WHERE: fix the broken filters*)?

## Learn more

- [SQLite: CREATE INDEX](https://www.sqlite.org/lang_createindex.html)
- [SQLite: EXPLAIN QUERY PLAN](https://www.sqlite.org/eqp.html)
- [Use The Index, Luke: a guide to SQL indexing](https://use-the-index-luke.com/)
