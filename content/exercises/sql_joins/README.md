# JOINs: services, teams and who to page

A deploy failed. Which team owns that service, and which Slack channel do you
alert? The answer is spread over three tables. A *JOIN* puts the pieces back
together.

**New here:** several related tables, `JOIN … ON`, table aliases, `table.column`, joining three tables.

## Learn

### Why the data is split up

Until now every deploy row held its service name as text. Real schemas store
each fact **once**: a `teams` table, a `services` table pointing at its team,
and a `deploys` table pointing at its service. Open `testdata/schema.sql`
and `testdata/data.sql`:

```text
teams                         services                           deploys
id | name     | slack_channel  id | name     | team_id | tier     id | service_id | version | ...
---+----------+-------------   ---+----------+---------+------    ---+------------+---------
 1 | payments | #payments-…     2 | checkout |       1 | crit…     4 |          2 | v2.7.1
 3 | search   | #search-team    4 | search   |       3 | stan…     6 |          4 | v3.1.0
```

Deploy 4 has `service_id` 2, so it's a `checkout` deploy; `checkout` has
`team_id` 1, so `payments` owns it. A column that holds another table's `id`
is a *foreign key*: it points at a row over there.

### JOIN … ON

`JOIN` glues each row of one table to the rows of another where the `ON`
condition is true:

```sql
SELECT services.name, services.tier, oncall.engineer
FROM services
JOIN oncall ON oncall.service_id = services.id;
```

For each service, the database finds the `oncall` row whose `service_id`
matches the service's `id`, and returns one combined row.

`JOIN` on its own means *inner join*: rows with **no** match on the other
side are dropped. A service with nobody on call doesn't appear above. (The
next exercise keeps them.)

```text
services            oncall               JOIN result
1 api        ───▶   1 priya       ───▶   api    | priya
2 checkout   ───▶   2 marco       ───▶   checkout | marco
4 search     ───▶   (no row)             (dropped)
```

### Say which table a column comes from

`services` and `teams` both have a column called `name`. Write `name` alone
and SQLite stops with `ambiguous column name: name`. Write
`table.column` instead. *Aliases* keep that short, and `AS` on a column
gives the output a clear name:

```sql
SELECT s.name AS service, s.tier, o.engineer
FROM services AS s
JOIN oncall AS o ON o.service_id = s.id
WHERE s.tier = 'critical';
```

> **Why is it written like this?** SQL requires the table prefix only when a
> name is ambiguous, but many teams prefix every column in a join so readers
> never have to guess. One-letter aliases (`s`, `t`, `d`) are a convention
> for short queries; the `AS` before a table alias is optional. Writing the
> `ON` condition as `other.foreign_key = this.id` is a style choice.

### Joining three tables

Each extra table is one more `JOIN … ON` line. The order of the steps:
`FROM`, all the `JOIN`s, then `WHERE`, `GROUP BY`, `ORDER BY` as before.

```sql
SELECT d.version, s.name AS service, t.slack_channel
FROM deploys AS d
JOIN services AS s ON s.id = d.service_id
JOIN teams AS t ON t.id = s.team_id
WHERE d.environment = 'staging';
```

`WHERE`, `GROUP BY` and aggregates work on the joined rows exactly as they did
on one table.

## Your task

Build the lookups the incident bot needs. Each task names the columns
(use `AS`) and the order of the rows.

### 1. Show which team owns each service

**Goal:** `1_service_teams.sql` returns `service` (the service's name) and
`team` (its team's name) for every service that has a team, A to Z by
service.

```text
service       | team
api           | platform
billing       | payments
checkout      | payments
notifications | platform
search        | search
```

`reports` has no team yet, so an inner join leaves it out.

**Think it through:**
- Which column in `services` points at `teams`, and which column does it
  match?
- Both tables have `name`: how do you say which one you mean?

### 2. Name the service of each failed deploy

**Goal:** `2_failed_deploys.sql` returns `id` (the deploy's id), `service`
and `version` for every deploy with status `failed`, by deploy id.

```text
id | service  | version
4  | checkout | v2.7.1
6  | search   | v3.1.0
10 | billing  | v1.0.0
14 | checkout | v2.8.0
```

**Think it through:**
- Start `FROM deploys`. Which table holds the name you need?
- Both tables have an `id` column. Which one goes in the first column?

### 3. Find the Slack channel to alert

**Goal:** `3_who_to_page.sql` returns `id`, `service`, `status` and
`slack_channel` for every deploy whose status is **not** `succeeded`, by
deploy id.

```text
id | service  | status      | slack_channel
4  | checkout | failed      | #payments-oncall
6  | search   | failed      | #search-team
10 | billing  | failed      | #payments-oncall
11 | billing  | rolled_back | #payments-oncall
14 | checkout | failed      | #payments-oncall
```

**Think it through:**
- The path is deploy → service → team. How many `JOIN` lines is that?

### 4. Count failed deploys per team

**Goal:** `4_failures_per_team.sql` returns `team` and `failures` (its number
of deploys with status `failed`) for each team with at least one, A to Z by
team.

```text
team     | failures
payments | 3
search   | 1
```

**Think it through:**
- Join first, filter the failed rows, then group: by what?

## Learn more

- [SQLite: the FROM clause and joins](https://www.sqlite.org/lang_select.html#fromclause)
- [SQLBolt, lesson 6: multi-table queries with JOINs](https://sqlbolt.com/lesson/select_queries_with_joins)
- [Join (SQL) on Wikipedia](https://en.wikipedia.org/wiki/Join_%28SQL%29)
