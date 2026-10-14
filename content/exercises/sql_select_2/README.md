# SELECT and WHERE: fix the broken filters

A teammate's deploy dashboard runs four queries, and each one is quietly
wrong: no error, just the wrong rows. These are the `WHERE` bugs you'll meet
most often in real reports.

**New here:** `LIKE` and `%`, `AND`/`OR` precedence, parentheses, `BETWEEN`, comparing dates stored as text.

## Learn

### Recap

As in *SELECT and WHERE: find the failed deploys*: `SELECT columns FROM
table WHERE condition`. Text goes in single quotes; `=` compares exactly;
`AND`, `OR` and `IN (...)` combine conditions. The table is the same
`deploys` table, with two more rows: open `testdata/data.sql`.

When a test fails it prints the rows your query returned and the rows it
wanted. Compare them: the extra or missing rows point at the bug.

### Exact means exact

`=` on text compares every character, including case. `'production'` and
`'Production'` are different strings. Check the real values in the data
before you trust a filter.

### AND binds tighter than OR

SQL evaluates `AND` before `OR`, the way maths does `*` before `+`. So

```sql
WHERE duration_s > 300 AND service = 'checkout' OR service = 'billing'
```

means `(slow AND checkout) OR billing`: every billing deploy sneaks in, fast
or slow. Parentheses say what you mean:

```sql
WHERE duration_s > 300 AND (service = 'checkout' OR service = 'billing')
```

To find a bug like this, write the condition out with brackets the way SQL
reads it, then check one of the extra rows against it.

> **Why is it written like this?** The precedence is a rule of SQL: you can't
> change it. Parentheses are your choice, and many teams add them whenever
> `AND` and `OR` are mixed, even when not needed, so the reader doesn't have
> to remember the rule. `IN (...)` avoids the problem entirely.

### LIKE matches patterns

`=` never treats any character as special. To match a pattern, use `LIKE`:

| Pattern | Matches |
|---|---|
| `'api-%'` | text starting with `api-` (`%` is any run of characters, even none) |
| `'%.log'` | text ending with `.log` |
| `'%timeout%'` | text containing `timeout` anywhere |
| `'v1._'` | `v1.` then exactly one character (`_` is one character) |

```sql
SELECT service, version
FROM deploys
WHERE version LIKE 'v3.%';
```

In SQLite, `LIKE` ignores case for English letters; Postgres's doesn't.

### Dates as text, and BETWEEN

SQLite has no date type. Here times are stored as text like
`'2026-03-04 11:15:00'`. That format sorts and compares correctly as plain
text, because the biggest unit comes first.

`a BETWEEN x AND y` means `a >= x AND a <= y`, both ends included. The trap:
`'2026-03-04'` is a *shorter string*, and every time on March 4th is
**greater** than it:

```text
'2026-03-04'           <  '2026-03-04 00:00:01'   (shorter string, same start)
'2026-03-04 11:15:00'  >  '2026-03-04'
```

For a range of whole days, use a *half-open range*: from the first day,
up to but **not** including the day after the last one:

```sql
WHERE started_at >= '2026-03-02'
  AND started_at <  '2026-03-03'   -- all of March 2nd
```

## Your task

Fix each query in place. Keep the columns as they are; only the `WHERE` line
needs to change. Rows may come back in any order.

### 1. Fix the staging filter

**Goal:** `1_staging_deploys.sql` returns the `service` and `version` of all
8 staging deploys. Right now it returns nothing.

```text
api      | v1.4.0
checkout | v2.7.1
search   | v3.1.0
...
billing  | v1.1.0-rc.1
```

**Think it through:**
- Look at how `environment` is written in `testdata/data.sql`.
- Does `=` care about capital letters?

### 2. Fix the production problems report

**Goal:** `2_production_problems.sql` returns the `id`, `service` and
`status` of every **production** deploy whose status is `failed` or
`rolled_back`.

```text
4  | checkout | failed
11 | billing  | rolled_back
14 | checkout | failed
```

Right now it also returns deploy 15, a staging rollback.

**Think it through:**
- Which runs first, `AND` or `OR`? Write the condition out with brackets the
  way SQL reads it.
- Parentheses fix it; so does `IN`.

### 3. Fix the release candidate search

**Goal:** `3_release_candidates.sql` returns the `service` and `version` of
every deploy whose version contains `-rc`.

```text
api     | v1.5.0-rc.1
billing | v1.1.0-rc.1
```

**Think it through:**
- The pattern `'%-rc%'` is right. What does `=` do with a `%`?

### 4. Fix the date range

**Goal:** `4_mid_week.sql` returns the `id`, `service` and `started_at` of
every deploy that started on **March 3rd or March 4th** 2026, whatever the
time of day.

```text
5  | billing  | 2026-03-03 08:45:00
6  | search   | 2026-03-03 09:20:00
7  | search   | 2026-03-03 10:01:00
8  | checkout | 2026-03-03 14:40:00
9  | api      | 2026-03-04 09:00:00
10 | billing  | 2026-03-04 11:15:00
11 | billing  | 2026-03-04 16:50:00
```

Right now the March 4th deploys are missing.

**Think it through:**
- Is `'2026-03-04 09:00:00'` less than or equal to `'2026-03-04'`?
- What is the first moment that is *not* in the range?

## Learn more

- [SQLite: expressions, including LIKE and BETWEEN](https://www.sqlite.org/lang_expr.html)
- [SQLite: operator precedence](https://www.sqlite.org/lang_expr.html#operators_and_parse_affecting_attributes)
- [SQLBolt, lesson 3: filtering with LIKE](https://sqlbolt.com/lesson/select_queries_with_constraints_pt_2)
