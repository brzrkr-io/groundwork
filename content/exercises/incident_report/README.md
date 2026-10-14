# Practice: incident report queries

It's Monday, and the head of engineering wants the incident review: what's
still open, how fast each service recovers, which postmortems are missing,
and which teams are drowning. Then you close the worst incident properly.
Every query uses ideas from this module; nothing here is new.

**New here:** nothing; practice with `JOIN`, `LEFT JOIN`, `NULL`, `GROUP BY`, `HAVING`, `ORDER BY` and transactions.

## Learn

### The tables

Open `testdata/schema.sql` and `testdata/data.sql`:

```text
teams ◀── services ◀── incidents ◀── postmortems
 id        id           id            incident_id
 name      name         service_id    author
           team_id      severity      published_at
                        title
                        opened_at
                        resolved_at          (NULL while open)
                        minutes_to_resolve   (NULL while open)
```

`severity` is `sev1` (worst), `sev2` or `sev3`. As text, `'sev1'` sorts
before `'sev2'`, so "worst first" is plain ascending order.

### Recap

| Need | Tool | From |
|---|---|---|
| combine tables | `JOIN … ON` | *JOINs: services, teams and who to page* |
| keep rows with no match | `LEFT JOIN`, then `IS NULL` | *LEFT JOIN: services with no on-call* |
| count only real matches | `COUNT(col)`, not `COUNT(*)` | *LEFT JOIN: services with no on-call* |
| missing values | `IS NULL`, `IS NOT NULL`; `AVG` skips `NULL` | *NULL: fix the missing-value bugs* |
| per-group numbers | `GROUP BY`, `HAVING` | *GROUP BY and HAVING: fix the flaky-service report* |
| a month of timestamps | `>=` the first day, `<` the next month | *SELECT and WHERE: fix the broken filters* |
| sort on two keys | `ORDER BY a DESC, b` | *ORDER BY and LIMIT: the latest deploys* |
| all or nothing | `BEGIN; … COMMIT;` | *INSERT, UPDATE and DELETE: record a release* |

The tests check column names (use `AS`) and the order of the rows, and show
`NULL` as `NULL`.

## Your task

Write one query (or, for task 5, one transaction) per file.

### 1. List the open incidents

**Goal:** `1_open_incidents.sql` returns `id`, `service`, `severity` and
`opened_at` of every incident that is **still open**, worst severity first,
and oldest first within a severity.

```text
id | service  | severity | opened_at
7  | checkout | sev1     | 2026-03-06 10:05:00
6  | api      | sev3     | 2026-03-06 08:00:00
8  | billing  | sev3     | 2026-03-06 11:20:00
```

**Think it through:**
- Which column says an incident is still open?
- Which table holds the service's name?

### 2. Report time to resolve per service

**Goal:** `2_time_to_resolve.sql` returns, for **every** service, `service`,
`incidents` (its number of incidents, open or not) and `avg_minutes` (the
average `minutes_to_resolve` of its resolved incidents, rounded to one
decimal place), A to Z.

```text
service       | incidents | avg_minutes
api           | 3         | 57.5
billing       | 2         | 70
checkout      | 3         | 30
notifications | 0         | NULL
search        | 1         | 90
```

The test prints `70.0` as `70`.

**Think it through:**
- `notifications` has no incidents. Which join keeps it, and which `COUNT`
  gives 0?
- Do open incidents need excluding from the average, or does `AVG` handle
  their `NULL`s?

### 3. Find the missing postmortems

**Goal:** `3_missing_postmortems.sql` returns `id`, `service` and `severity`
of every **resolved** `sev1` or `sev2` incident that has **no** postmortem,
by id.

```text
id | service  | severity
2  | api      | sev2
5  | checkout | sev2
```

**Think it through:**
- Three conditions in the `WHERE`. Which one needs a `LEFT JOIN` first?

### 4. Rank the noisiest teams in March

**Goal:** `4_noisiest_teams.sql` returns `team` and `incidents` (how many of
its services' incidents were **opened in March 2026**) for each team with
**at least 2**, most incidents first, then A to Z by team.

```text
team     | incidents
payments | 5
platform | 2
```

Incident 9 was in February, so it doesn't count for platform.

**Think it through:**
- The path is incident → service → team: two joins.
- Which condition filters rows, and which filters groups?

### 5. Close incident 7 with its postmortem

**Goal:** `5_close_incident.sql` does two things inside one transaction
(`BEGIN;` … `COMMIT;`), in this order:

1. set incident 7's `resolved_at` to `2026-03-06 11:02:00` and its
   `minutes_to_resolve` to `57`;
2. add its postmortem: `incident_id` 7, `author` `marco`, `published_at`
   `2026-03-09 10:00:00`.

The test also runs your file against a database where adding the postmortem
fails, and checks that incident 7 stays open.

**Think it through:**
- One `UPDATE` setting two columns, one `INSERT`. Which `WHERE` keeps the
  update to incident 7?

## Learn more

- [SQLite: SELECT](https://www.sqlite.org/lang_select.html)
- [Google SRE book: Postmortem Culture](https://sre.google/sre-book/postmortem-culture/)
