# SELECT and WHERE: find the failed deploys

Deploy history, incidents, users, feature flags: most of the data an ops
team cares about lives in a *database*, and you ask it questions with SQL
(*Structured Query Language*). Here you write your first queries against a
table of deploys.

**New here:** SQL, tables, `SELECT`, `FROM`, `WHERE`, `AND`, `OR`, `IN`, comparison operators.

## Learn

### How these exercises work

This exercise is SQL, not Go. Each task is one `.sql` file
(`1_all_deploys.sql`, `2_failed_deploys.sql`, …). You write one query under
its `TODO`.

The test builds a fresh database in memory with *SQLite*, a small database
that runs inside the test program. It loads `testdata/schema.sql` (the table)
and `testdata/data.sql` (the rows), then runs your query. Save the file and
the result appears: when a test fails, it prints every row your query
returned next to the rows it wanted. Use that to explore.

Open `testdata/data.sql` now: it's the whole table, 14 deploys.

### Tables, rows and columns

A *table* is like a spreadsheet with strict rules. Each *column* has a name
and a type; each *row* is one record. Here are a few rows of `deploys`:

```text
id | service  | version | environment | status    | started_at          | duration_s
---+----------+---------+-------------+-----------+---------------------+-----------
 1 | api      | v1.4.0  | staging     | succeeded | 2026-03-02 09:12:00 |         95
 4 | checkout | v2.7.1  | production  | failed    | 2026-03-02 13:02:00 |        610
 5 | billing  | v0.9.3  | production  | succeeded | 2026-03-03 08:45:00 |         88
```

`id` numbers each row. `duration_s` is a whole number of seconds; the rest is
text.

### SELECT … FROM

`SELECT` picks columns, `FROM` names the table:

```sql
SELECT version, status
FROM deploys;
```

This returns two columns, in the order you list them, for **every** row.
`SELECT *` means "all columns", handy for a quick look but not for the tasks:
the tests check exactly the columns each task asks for.

> **Why is it written like this?** SQL requires `SELECT` before `FROM`, and
> commas between columns. Keywords work in any case (`select` = `SELECT`);
> writing them in capitals is a convention so they stand out from names.
> Line breaks are your choice: the query above could be one line. The `;`
> ends a statement; with one query per file it's optional, but a good habit.

### WHERE keeps some rows

`WHERE` is a condition checked against each row. Only rows where it's true
come back:

```sql
SELECT service, started_at
FROM deploys
WHERE environment = 'staging';
```

| Operator | Means | Example |
|---|---|---|
| `=` | equal (one `=`, not `==`) | `status = 'failed'` |
| `!=` or `<>` | not equal | `service != 'api'` |
| `<` `>` `<=` `>=` | less / greater | `duration_s >= 60` |

Text goes in **single** quotes: `'staging'`. Numbers don't: `60`. Text
comparison is exact, so `'Staging'` matches nothing here.

> **Why is it written like this?** In SQL, single quotes make text and double
> quotes mean a *name* (a column or table). `WHERE env = "staging"` asks to
> compare two columns. SQLite guesses what you meant when no such column
> exists; Postgres refuses. Always use single quotes for text.

### Combining conditions

`AND` needs both sides true; `OR` needs at least one. `IN` is a short way to
write several `OR`s on the same column:

```sql
SELECT id, version
FROM deploys
WHERE service = 'search'
  AND duration_s < 60;

SELECT id, status
FROM deploys
WHERE environment IN ('staging', 'production');
-- same as: environment = 'staging' OR environment = 'production'
```

`NOT IN (...)` is the opposite of `IN`.

Mixing `AND` with `OR`? Put brackets around the `OR` part:
`WHERE (environment = 'staging' OR status = 'failed') AND duration_s > 60`.
Without them, SQL groups the conditions differently from how you read them,
as you'll see in *SELECT and WHERE: fix the broken filters*.

### Comments

`--` starts a comment that runs to the end of the line, like `//` in Go.
Each task file starts with comments describing the task.

## Your task

Your team wants answers about last week's deploys. Write one query per file.
Rows may come back in any order; the tests sort them before comparing.

### 1. List every deploy

**Goal:** in `1_all_deploys.sql`, return the `service`, `version` and
`environment` of every deploy, in that column order.

All 14 rows come back. The first few:

```text
api      | v1.4.0 | staging
api      | v1.4.0 | production
checkout | v2.7.1 | staging
```

**Think it through:**
- Which keyword lists the columns, and which one names the table?
- Do you need a `WHERE` when you want every row?

### 2. Find the failed deploys

**Goal:** in `2_failed_deploys.sql`, return the `service` and `version` of
every deploy whose `status` is `failed`.

Four rows:

```text
checkout | v2.7.1
search   | v3.1.0
billing  | v1.0.0
checkout | v2.8.0
```

**Think it through:**
- Which column holds `failed`, and is it text or a number?
- Which kind of quotes does text need?

### 3. Find slow production deploys

**Goal:** in `3_slow_production.sql`, return the `id`, `service` and
`duration_s` of every **production** deploy that took **more than 300**
seconds.

```text
4  | checkout | 610
8  | checkout | 395
11 | billing  | 720
14 | checkout | 455
```

Deploy 10 took 302 seconds but went to staging, so it's not here.

**Think it through:**
- There are two conditions. Which word makes a row pass only if both hold?
- "More than 300": does 300 itself count?

### 4. Find the payments team's problem deploys

**Goal:** in `4_payments_problems.sql`, return the `id`, `service` and
`status` of every deploy of `checkout` or `billing` whose status is anything
other than `succeeded`.

```text
4  | checkout | failed
10 | billing  | failed
11 | billing  | rolled_back
14 | checkout | failed
```

**Think it through:**
- One condition picks two services: which operator lists several values?
- "Anything other than": which operator means not equal?

## Learn more

- [SQLite: SELECT](https://www.sqlite.org/lang_select.html)
- [SQLBolt, lessons 1–3](https://sqlbolt.com/lesson/select_queries_introduction): interactive SELECT practice in the browser
- [SQL (Wikipedia)](https://en.wikipedia.org/wiki/SQL)
