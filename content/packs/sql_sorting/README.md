# ORDER BY and LIMIT: the latest deploys

"What went out last?" is the first question in almost every incident. Here
you sort query results, keep only the top few rows, and remove duplicates.

**New here:** `ORDER BY`, `ASC`, `DESC`, sorting on several columns, `LIMIT`, `DISTINCT`.

## Learn

### Recap

As in *SELECT and WHERE: find the failed deploys*: one query per `.sql`
file, run against the `deploys` table in `testdata/data.sql`. A failing test
prints the rows you returned and the rows it wanted.

### Rows have no order until you ask

A table is a *set* of rows, not a list. Without `ORDER BY`, the database may
return them in any order, and that order can change when the data grows or
an index is added. If the order matters, say so.

The tests check this. They run your query twice, the second time with
SQLite told to read tables backwards. A query without `ORDER BY` may look
right on the first run, but the second run gives it away.

### ORDER BY

`ORDER BY` comes after `WHERE` and sorts the result:

```sql
SELECT service, duration_s
FROM deploys
WHERE environment = 'staging'
ORDER BY duration_s;        -- smallest first
```

| You write | Order |
|---|---|
| `ORDER BY duration_s` or `ORDER BY duration_s ASC` | ascending: smallest, oldest, A first |
| `ORDER BY duration_s DESC` | descending: largest, newest, Z first |

Text sorts alphabetically, numbers by size, and the timestamps here sort by
time (the biggest unit comes first in `'2026-03-04 11:15:00'`).

> **Why is it written like this?** SQL fixes the order of the parts: `SELECT`,
> `FROM`, `WHERE`, `ORDER BY`, `LIMIT`. Put them in another order and you get
> a syntax error. `ASC` is the default, so most people leave it out.

### Sorting on several columns

List several columns, separated by commas. The second one only breaks ties
in the first, and each column gets its own `ASC` or `DESC`:

```sql
SELECT environment, started_at
FROM deploys
ORDER BY environment, started_at DESC;
-- production rows first (p < s), newest first within each environment
```

### LIMIT keeps the first few

`LIMIT n` goes last and keeps only the first `n` rows of the result. With
`ORDER BY ... DESC`, that's "the top n":

```sql
SELECT id, duration_s
FROM deploys
ORDER BY duration_s DESC
LIMIT 5;                    -- the five slowest deploys
```

`LIMIT` without `ORDER BY` gives you *some* `n` rows, not the first or last.

### DISTINCT removes duplicates

`SELECT DISTINCT` drops rows that are exact copies of an earlier row:

```sql
SELECT DISTINCT environment
FROM deploys;
-- staging
-- production   (each once, though there are 14 deploys)
```

## Your task

The incident channel wants a few quick answers about recent deploys. The
order of the rows is now part of each answer.

### 1. Sort the deploys by time

**Goal:** `1_deploy_timeline.sql` returns the `id`, `service` and
`started_at` of every deploy, **oldest first**.

```text
1  | api      | 2026-03-02 09:12:00
2  | api      | 2026-03-02 10:05:00
3  | checkout | 2026-03-02 11:30:00
...
14 | checkout | 2026-03-06 10:00:00
```

**Think it through:**
- Which column holds the time?
- Is oldest first ascending or descending?

### 2. Show the three latest production deploys

**Goal:** `2_latest_production.sql` returns the `id`, `service` and `version`
of the **three** most recent production deploys, **newest first**.

```text
14 | checkout | v2.8.0
13 | api      | v1.4.1-hotfix
12 | search   | v3.1.1
```

**Think it through:**
- Which part keeps only production deploys, and where does it go?
- Sort first, then cut: in which order do `ORDER BY` and `LIMIT` appear?

### 3. Sort by service, then slowest first

**Goal:** `3_slowest_per_service.sql` returns the `service` and `duration_s`
of every deploy, sorted by service **A to Z**, and within each service by
duration, **slowest first**.

```text
api      | 140
api      | 101
api      | 95
api      | 77
billing  | 720
...
search   | 45
```

**Think it through:**
- Which column decides the order first, and which one only breaks ties?
- Which of the two needs `DESC`?

### 4. List each production service once

**Goal:** `4_production_services.sql` returns one column, `service`: each
service that has had at least one production deploy, **once**, A to Z.

```text
api
billing
checkout
search
```

**Think it through:**
- On its own, the query lists `checkout` three times. Which word removes the copies?
- You still need `WHERE` and `ORDER BY`.

## Learn more

- [SQLite: SELECT, the ORDER BY and LIMIT clauses](https://www.sqlite.org/lang_select.html#the_order_by_clause)
- [SQLBolt, lesson 4: filtering and sorting](https://sqlbolt.com/lesson/filtering_sorting_query_results)
