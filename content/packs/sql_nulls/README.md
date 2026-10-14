# NULL: fix the missing-value bugs

Real tables have holes: a deploy that hasn't finished, a field an old script
never filled in. SQL marks a hole with `NULL`, and `NULL` breaks the rules
you're used to. Four reports here are wrong because of it.

**New here:** `NULL` and three-valued logic, `IS NULL`, `IS NOT NULL`, `COALESCE`, how aggregates treat `NULL`, `||`.

## Learn

### NULL means "unknown"

`NULL` isn't zero, and it isn't an empty string. It means *no value here*, or
"unknown". Open `testdata/data.sql`: `finished_at` is `NULL` while a deploy is
running, `environment` and `duration_s` are `NULL` where the old script didn't
record them, and `deployed_by` is `NULL` when CI did the deploy.

In `testdata/schema.sql`, columns with `NOT NULL` can never hold `NULL`; the
others can.

### Comparisons with NULL are unknown

SQL asks: is `NULL = 'production'`? It can't know, so the answer is neither
true nor false but *unknown* (itself `NULL`). That's *three-valued logic*:

| Expression | Result |
|---|---|
| `NULL = 'production'` | unknown |
| `NULL != 'production'` | unknown |
| `NULL = NULL` | unknown (two unknowns aren't known to be equal) |
| `NULL IS NULL` | **true** |
| `'staging' IS NULL` | false |

`WHERE` keeps a row only when the condition is **true**. Unknown counts as
"no", so `NULL` rows silently fall out of both `= x` **and** `!= x`.

To test for `NULL`, use `IS NULL` or `IS NOT NULL`:

```sql
SELECT id
FROM deploys
WHERE deployed_by IS NOT NULL;   -- deploys a person started
```

> **Why is it written like this?** `IS NULL` is SQL's only way to ask "is
> this missing?"; `= NULL` is valid syntax but never true, so the database
> won't warn you. To include the unknowns in a "not equal" filter, say so:
> `col IS NULL OR col != 'x'`. SQLite also has `col IS NOT 'x'`, which
> treats `NULL` as an ordinary value (Postgres writes it
> `col IS DISTINCT FROM 'x'`).

### NULL spreads through expressions

Any arithmetic or text operation with a `NULL` gives `NULL`. `||` joins text
in SQL (like `+` on Go strings):

```sql
SELECT 'v' || '1.4'      -- 'v1.4'
SELECT 'v' || NULL       -- NULL
SELECT 60 + NULL         -- NULL
```

`COALESCE(a, b, ...)` returns the first argument that isn't `NULL`. It's the
standard way to fill in a default:

```sql
SELECT COALESCE(environment, 'unknown') AS env
FROM deploys;
-- 'staging', 'production' or 'unknown'
```

### Aggregates skip NULL

`SUM`, `AVG`, `MIN`, `MAX` and `COUNT(col)` ignore `NULL`s. Only `COUNT(*)`
counts every row:

```text
duration_s:  60, 90, NULL

COUNT(*)          3
COUNT(duration_s) 2
SUM(duration_s)   150
AVG(duration_s)   75.0       (150 / 2, not 150 / 3)
```

That's usually what you want: an unknown duration shouldn't count as 0
seconds. An aggregate over only `NULL`s returns `NULL`.

## Your task

Fix each query in its file. Keep the column names and the order of the rows;
the test shows a `NULL` as `NULL`.

### 1. Fix the running deploys query

**Goal:** `1_running_deploys.sql` returns `id`, `service` and `started_at` of
every deploy with **no** `finished_at`, by id.

```text
id | service | started_at
8  | billing | 2026-03-06 16:58:00
9  | api     | 2026-03-06 17:02:00
```

**Think it through:**
- What does `finished_at = NULL` give for each row?

### 2. Fix the non-production filter

**Goal:** `2_not_production.sql` returns `id`, `service` and `environment` of
every deploy that is **not known to be** production: staging deploys **and**
deploys whose `environment` is `NULL`, by id.

```text
id | service  | environment
1  | api      | staging
3  | checkout | NULL
6  | search   | staging
7  | search   | NULL
8  | billing  | staging
```

Right now 3 and 7 are missing.

**Think it through:**
- Is `NULL != 'production'` true?
- How do you add "or there's no environment at all"?

### 3. Fix the deployer labels

**Goal:** `3_deployer_labels.sql` returns `id` and `label` for every deploy,
by id. `label` is `deployed by ` followed by `deployed_by`, or `deployed by
automation` when `deployed_by` is `NULL`.

```text
id | label
1  | deployed by priya
...
3  | deployed by automation
...
10 | deployed by li
```

Right now the pipeline's deploys get a `NULL` label.

**Think it through:**
- What does `'deployed by ' || NULL` give?
- Which function swaps a `NULL` for a default, and where does it go?

### 4. Fix the duration report

**Goal:** `4_duration_report.sql` returns, for each service, `service`,
`deploys` (**all** its deploys) and `avg_duration_s` (the average of its
**recorded** durations, rounded to one decimal place), A to Z.

```text
service  | deploys | avg_duration_s
api      | 3       | 117.5
billing  | 2       | NULL
checkout | 3       | 425.3
search   | 2       | 48.5
```

`billing` has no recorded durations, so its average is unknown: `NULL`.

**Think it through:**
- Which `COUNT` counts every row, and which one skips `NULL`s?
- The hand-made average divides by every row. Which aggregate already
  divides by the recorded ones only?

## Learn more

- [SQLite: NULL handling in SQLite](https://www.sqlite.org/nulls.html)
- [SQLite: core functions, including coalesce()](https://www.sqlite.org/lang_corefunc.html#coalesce)
- [Null (SQL) on Wikipedia](https://en.wikipedia.org/wiki/Null_%28SQL%29)
