# Aggregates and GROUP BY: failures per service

"Which service fails most?" and "how long does a deploy take?" are questions
about many rows at once. *Aggregate functions* boil rows down to one number,
and `GROUP BY` does that once per service, per day, per anything.

**New here:** `COUNT`, `SUM`, `AVG`, `MIN`, `MAX`, `ROUND`, `AS`, `GROUP BY`, `date()`.

## Learn

### Recap

Same `deploys` table as in *ORDER BY and LIMIT: the latest deploys* (see
`testdata/data.sql`). `WHERE` filters rows, `ORDER BY` sorts them.

### Aggregates: many rows in, one value out

| Function | Returns |
|---|---|
| `COUNT(*)` | the number of rows |
| `SUM(col)` | the total of a number column |
| `AVG(col)` | the average (a decimal number) |
| `MIN(col)` / `MAX(col)` | the smallest / largest value (works on text too) |

Without `GROUP BY`, an aggregate turns the whole table into **one** row:

```sql
SELECT COUNT(*), SUM(duration_s)
FROM deploys
WHERE service = 'api';
-- 4 | 413
```

`WHERE` runs first, so only the `api` rows are counted and summed.

### AS names a column

A computed column gets an ugly name like `COUNT(*)`. `AS` gives it a real one,
which reports and Go code can rely on:

```sql
SELECT MAX(duration_s) AS longest_s
FROM deploys;
```

> **Why is it written like this?** `AS` is part of SQL; the names are your
> choice. *snake_case* (`longest_s`) is the usual convention for column names,
> and a unit in the name (`_s`) saves a reader from guessing. The tests check
> the names each task asks for.

### ROUND

`ROUND(x, 1)` rounds to one decimal place: `ROUND(134.1666, 1)` is `134.2`.
Functions nest: `ROUND(AVG(duration_s), 1)`.

### GROUP BY: one result row per group

`GROUP BY` sorts the rows into groups that share a value, then runs the
aggregates once per group:

```text
deploys (filtered)            GROUP BY status         result
api      succeeded  ─┐
search   succeeded  ─┼──▶ succeeded: 3 rows ──▶  succeeded | 3
billing  succeeded  ─┘
checkout failed     ─┬──▶ failed: 2 rows    ──▶  failed    | 2
search   failed     ─┘
```

```sql
SELECT status, COUNT(*) AS deploys
FROM deploys
GROUP BY status;
```

Every column in the `SELECT` should be either a `GROUP BY` column or inside
an aggregate. A plain column that isn't grouped has no single value per
group: Postgres rejects it, and SQLite quietly picks the value from one of
the group's rows.

The parts go in this order: `SELECT`, `FROM`, `WHERE`, `GROUP BY`,
`ORDER BY`, `LIMIT`. `ORDER BY` may use a name made with `AS`.

### Grouping by a computed value

You can group by an expression, such as part of a timestamp. `date(t)`
returns the `YYYY-MM-DD` part of a SQLite timestamp:

```sql
SELECT date('2026-03-04 11:15:00');
-- 2026-03-04
```

Give it a name with `AS`, then `GROUP BY` that name.

## Your task

Build the numbers for the weekly deploy review. Each task says the column
names to use and the order of the rows.

### 1. Summarize every deploy

**Goal:** `1_deploy_summary.sql` returns **one** row with three columns:
`deploys` (how many deploys there are), `slowest_s` (the largest
`duration_s`) and `first_started` (the earliest `started_at`).

```text
deploys | slowest_s | first_started
14      | 720       | 2026-03-02 09:12:00
```

**Think it through:**
- Which aggregate gives each of the three values?
- What gives a column the name the task asks for?

### 2. Count failures per service

**Goal:** `2_failures_per_service.sql` returns `service` and `failures` (the
number of failed deploys) for each service with at least one failed deploy,
A to Z by service.

```text
service  | failures
billing  | 1
checkout | 2
search   | 1
```

**Think it through:**
- Which rows should be counted at all? Filter them before grouping.
- What do you group by to get one row per service?

### 3. Average the duration per environment

**Goal:** `3_average_duration.sql` returns, for each environment,
`environment`, `deploys` (its number of deploys) and `avg_duration_s` (its
average `duration_s`, rounded to **one** decimal place), A to Z by
environment.

```text
environment | deploys | avg_duration_s
production  | 8       | 318.3
staging     | 6       | 134.2
```

**Think it through:**
- Two aggregates in one query, both computed per group.
- Which function goes on the outside: `ROUND` or `AVG`?

### 4. Count deploys per day

**Goal:** `4_deploys_per_day.sql` returns `day` (the `YYYY-MM-DD` date of
`started_at`) and `deploys` (how many deploys started that day) for every
day with deploys, **newest day first**.

```text
day        | deploys
2026-03-06 | 1
2026-03-05 | 2
2026-03-04 | 3
2026-03-03 | 4
2026-03-02 | 4
```

**Think it through:**
- What do you group by: the full timestamp, or only its date?
- Can you reuse the name you gave the day in `GROUP BY` and `ORDER BY`?

## Learn more

- [SQLite: aggregate functions](https://www.sqlite.org/lang_aggfunc.html)
- [SQLite: date and time functions](https://www.sqlite.org/lang_datefunc.html)
- [SQLBolt, lesson 10: queries with aggregates](https://sqlbolt.com/lesson/select_queries_with_aggregates)
