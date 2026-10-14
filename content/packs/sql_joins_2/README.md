# Left joins: services with no on-call

The scariest rows are the ones that *aren't there*: a service nobody is on
call for, a team that never deployed. An inner join hides them. A *LEFT
JOIN* keeps them, so you can find them.

**New here:** `LEFT JOIN`, `NULL` from a missing match, `IS NULL`, `COUNT(col)` vs `COUNT(*)`, `ON` vs `WHERE` in a left join.

## Learn

### Recap

As in *Joins: services, teams and who to page*: `JOIN other ON condition`
combines matching rows, and an inner join drops rows with no match. The
tables are the same: `teams`, `services`, `oncall`, `deploys` (see
`testdata/`).

### LEFT JOIN keeps every row on the left

`LEFT JOIN` keeps **every** row of the table before it (the left side). When
a row has no match, the columns from the right side are filled with `NULL`,
SQL's marker for "no value":

```text
services (left)       teams (right)        LEFT JOIN result
api      team 2  ───▶  2 platform   ───▶   api      | platform
reports  no team ───▶  (no match)   ───▶   reports  | NULL
```

```sql
SELECT s.name AS service, t.name AS team
FROM services AS s
LEFT JOIN teams AS t ON t.id = s.team_id;
-- ... reports | NULL
```

| | `JOIN` (inner) | `LEFT JOIN` |
|---|---|---|
| left row with a match | kept | kept |
| left row with no match | dropped | kept, right side `NULL` |

The test prints a `NULL` as `NULL`.

### Finding what's missing

To find left rows with **no** match, left-join, then keep the rows where the
right side came back empty. Test for `NULL` with `IS NULL`:

```sql
SELECT s.name AS service
FROM services AS s
LEFT JOIN teams AS t ON t.id = s.team_id
WHERE t.id IS NULL;   -- services with no team
```

> **Why is it written like this?** SQL requires `IS NULL`: `= NULL` is never
> true, not even for a `NULL` (*NULL: fix the missing-value bugs* explains
> why). Checking the right table's `id` is a convention: a real row always
> has one, so `NULL` there can only mean "no match".

### COUNT(*) vs COUNT(column)

After a left join, a service with no deploys still has **one** row, with
`NULL`s on the right. `COUNT(*)` counts that row, giving 1. `COUNT(column)`
counts only rows where the column is **not** `NULL`, giving 0:

```text
service       d.id      COUNT(*)  COUNT(d.id)
notifications NULL      1         0
```

So count a right-side column, like `COUNT(d.id)`.

Group by the left table's `id` (`GROUP BY s.id`): two services could share a
name, but never an id. You can still select `s.name`, although *Aggregates
and GROUP BY: failures per service* said every selected column should be
grouped: `id` is the primary key, so each group holds exactly one service
and so exactly one name. Postgres allows this for the same reason.

### ON vs WHERE in a left join

With a left join, *where* you put a condition changes the answer:

- in `ON`: it decides which right rows **match**. Left rows still all come
  back.
- in `WHERE`: it runs **after** the join and removes whole rows, including
  the left rows that only had `NULL`s. Your left join quietly turns into an
  inner join.

```sql
-- every team, with its critical services (NULL if none)
SELECT t.name AS team, s.name AS service
FROM teams AS t
LEFT JOIN services AS s
    ON s.team_id = t.id
    AND s.tier = 'critical';
```

Move `s.tier = 'critical'` to a `WHERE` and the `search` and `data` teams
vanish.

## Your task

The on-call lead wants to see the gaps. Each task names the columns and the
order of the rows.

### 1. List every service with its on-call engineer

**Goal:** `1_oncall_roster.sql` returns `service` and `engineer` for
**every** service, A to Z by service. A service with nobody on call shows
`NULL`.

```text
service       | engineer
api           | priya
billing       | marco
checkout      | marco
notifications | li
reports       | NULL
search        | NULL
```

**Think it through:**
- Which table must keep all its rows? Put it on the left.
- Which column links `oncall` to `services`?

### 2. Find the services nobody is on call for

**Goal:** `2_no_oncall.sql` returns one column, `service`: every service with
no row in `oncall`, A to Z.

```text
service
reports
search
```

**Think it through:**
- After a left join, what do the `oncall` columns hold for those services?
- Which operator tests for that?

### 3. Count deploys per service, including zero

**Goal:** `3_deploys_per_service.sql` returns `service` and `deploys` for
**every** service, A to Z. A service that never deployed shows `0`.

```text
service       | deploys
api           | 4
billing       | 3
checkout      | 4
notifications | 0
reports       | 0
search        | 3
```

**Think it through:**
- Which join keeps `notifications`, which has no deploys?
- `COUNT(*)` would give it 1. What do you count instead?

### 4. Fix the production deploy count

**Goal:** `4_production_deploys.sql` returns `service` and `deploys` (the
number of **production** deploys) for **every** service, A to Z, with `0`
for a service that has none.

```text
service       | deploys
api           | 2
billing       | 2
checkout      | 3
notifications | 0
reports       | 0
search        | 1
```

Right now `notifications` and `reports` are missing.

**Think it through:**
- When does `WHERE d.environment = 'production'` run: before or after the
  `NULL` rows are made? What does it do to them?
- Where can the condition go so it only decides which deploys match?

## Learn more

- [SQLite: the FROM clause and joins](https://www.sqlite.org/lang_select.html#fromclause)
- [SQLBolt, lesson 7: outer joins](https://sqlbolt.com/lesson/select_queries_with_outer_joins)
- [SQLBolt, lesson 8: a short note on NULLs](https://sqlbolt.com/lesson/select_queries_with_nulls)
