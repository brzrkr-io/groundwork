# GROUP BY and HAVING: fix the flaky-service report

The weekly "flaky services" report has four broken queries: one crashes, and
three return believable but wrong numbers. Wrong numbers are worse than an
error, because people act on them.

**New here:** `HAVING`, `WHERE` vs `HAVING`, `CASE WHEN`, `COUNT(DISTINCT col)`.

## Learn

### Recap

As in *Aggregates and GROUP BY: failures per service*: `GROUP BY service`
makes one result row per service, and `COUNT(*)`, `SUM`, `AVG`, `MIN`, `MAX`
are computed once per group. `AS` names a column. The data is the same
`deploys` table (`testdata/data.sql`).

### The order SQL runs a query in

You *write* `SELECT` first, but the database *runs* the parts in another
order:

```text
1. FROM      take the table
2. WHERE     keep some rows            (one row at a time)
3. GROUP BY  make groups
4. HAVING    keep some groups          (one group at a time)
5. SELECT    compute the output columns
6. ORDER BY  sort
7. LIMIT     cut
```

Most `GROUP BY` bugs come from doing a step at the wrong point in this list.

### HAVING filters groups

`WHERE` looks at one row at a time, before groups exist, so it can't use
`COUNT(*)` or `AVG(...)`: SQLite says `misuse of aggregate`. `HAVING` comes
after `GROUP BY` and filters whole groups by their aggregates:

```sql
SELECT status, COUNT(*) AS deploys
FROM deploys
GROUP BY status
HAVING COUNT(*) >= 2;
```

| | `WHERE` | `HAVING` |
|---|---|---|
| filters | rows | groups |
| runs | before grouping | after grouping |
| can use aggregates | no | yes |

A query can have both: `WHERE` decides which rows go into the groups,
`HAVING` decides which groups come out.

> **Why is it written like this?** SQL requires `HAVING` after `GROUP BY` and
> before `ORDER BY`. Repeating the aggregate (`HAVING COUNT(*) >= 2`) works
> in every database; SQLite also accepts the `AS` name there, but Postgres
> doesn't, so the repeat is the portable choice.

### Counting only some rows: CASE WHEN

A `WHERE` removes rows from **every** column. To count only some rows in one
column while another counts them all, turn each row into a 1 or a 0 and sum:

```sql
SELECT
    service,
    COUNT(*) AS deploys,
    SUM(CASE WHEN environment = 'staging' THEN 1 ELSE 0 END) AS staging_deploys
FROM deploys
GROUP BY service;
```

`CASE WHEN condition THEN a ELSE b END` is SQL's `if`: it gives `a` when the
condition is true, otherwise `b`.

### COUNT(DISTINCT col)

`COUNT(col)` counts rows where `col` has a value. `COUNT(DISTINCT col)`
counts **different** values: two deploys of `v1.0.0` count once.

```sql
SELECT COUNT(DISTINCT environment) AS environments
FROM deploys;
-- 2
```

### Dividing numbers

`duration_s / 60` divides two whole numbers, so SQL throws away the
remainder: `90 / 60` is `1`. Write `60.0` to get a decimal: `90 / 60.0` is
`1.5`. `AVG` already returns a decimal, but `60.0` makes your intent clear.

## Your task

Fix each query in its file. Keep the column names; the tests check them and
the order of the rows.

### 1. Fix the busy services query

**Goal:** `1_busy_services.sql` returns `service` and `deploys` for each
service with **more than 3** deploys, A to Z.

```text
service  | deploys
api      | 4
checkout | 4
```

Right now it stops with `misuse of aggregate: COUNT()`.

**Think it through:**
- Does "more than 3 deploys" describe a row or a group?
- Where does a group filter go, relative to `GROUP BY`?

### 2. Fix the failure counts

**Goal:** `2_failure_counts.sql` returns, for **every** service, `service`,
`deploys` (all its deploys) and `failures` (its deploys with status
`failed`), A to Z.

```text
service  | deploys | failures
api      | 4       | 0
billing  | 3       | 1
checkout | 4       | 2
search   | 3       | 1
```

Right now `api` is missing and both numbers count only failures.

**Think it through:**
- Which line throws away the rows that `deploys` should count?
- Once it's gone, how does `failures` count only failed rows?

### 3. Fix the version count

**Goal:** `3_versions_per_service.sql` returns `service` and `versions`, the
number of **different** versions each service has deployed, A to Z.

```text
service  | versions
api      | 3
billing  | 2
checkout | 3
search   | 2
```

`v1.4.0` of api went to staging and production: it counts once.

**Think it through:**
- What does `COUNT(version)` count: rows or values?

### 4. Fix the slow services report

**Goal:** `4_slow_production_services.sql` returns `service` and
`avg_minutes` for each service whose **production** deploys take more than
300 seconds **on average**. `avg_minutes` is that average in minutes,
rounded to one decimal place. Sort the rows A to Z.

```text
service  | avg_minutes
billing  | 6.7
checkout | 8.1
```

Right now billing shows `12`: the query only averages deploys over 300
seconds.

**Think it through:**
- "Over 300 seconds" should apply to the average, not to each deploy.
- Which condition stays in `WHERE`, and which one moves to `HAVING`?

## Learn more

- [SQLite: SELECT, how GROUP BY and HAVING work](https://www.sqlite.org/lang_select.html#generation_of_the_set_of_result_rows)
- [SQLite: CASE expressions](https://www.sqlite.org/lang_expr.html#the_case_expression)
- [SQLBolt, lesson 11: aggregates with HAVING](https://sqlbolt.com/lesson/select_queries_with_aggregates_pt_2)
