# Relational databases and SQL

**In one sentence:** a relational database keeps data in tables of rows and columns, enforces rules about what may go in them, and answers questions written in SQL, safely, even when many programs read and write at once.

**Why it matters:** deploy history, incidents, users, feature flags, the state of a CI system: almost every service you'll run on call has a database behind it. The 3 a.m. question "what changed, and who owns it?" is usually a SQL query, and the classic outages (a slow query that melts the database, a migration that locks a table, an `UPDATE` without a `WHERE`) are database problems.

**You'll learn:**

- Why a database beats a pile of files once data is shared, changing and important
- Tables, rows, columns, types, and how primary and foreign keys connect tables
- How to read and write a query: `SELECT`, `WHERE`, `ORDER BY`, `GROUP BY`, `JOIN`
- Why `NULL` breaks your intuition, and how indexes turn O(n) into O(log n)
- What transactions and ACID promise, and what normalization means in plain words
- How SQL injection works, why placeholders stop it, and how schemas change safely with migrations

## Why not just use files?

### A deploy log that outgrew itself

Your team starts by appending one line per deploy to `deploys.log`:

```text
2026-03-02 13:02:00 checkout v2.7.1 production failed
2026-03-03 14:40:00 checkout v2.7.2 production succeeded
```

It works for a month. Then the questions start, and each one hurts:

- **"How many checkout deploys failed this quarter?"** You write a script that reads the whole file. Next week someone wants it per team, so you write another one.
- **Two pipelines finish at the same moment.** Both append; on a bad day their lines interleave into garbage.
- **The deploy bot crashes halfway through writing a line.** The file now ends with half a record, and the next reader chokes on it.
- **Someone types `prodution`.** Nothing stops them, and every report silently misses that deploy.
- **The file reaches 40 million lines.** Finding one deploy means reading all of them.

### What a database adds

A *database* is a program whose whole job is storing data for other programs. Your code sends it requests; it does the careful work.

| Need | With files | With a database |
|---|---|---|
| ask a new question | write a new script | write one query |
| many writers at once | interleaved lines, lost updates | writes are coordinated for you |
| crash in the middle | half-written records | the change happens fully or not at all |
| bad values | anything goes | rules (*constraints*) reject them |
| find one record among millions | read everything | jump to it with an index |

Files are still right for many things: configs, logs that are only appended and grepped, build artifacts. Reach for a database when data is **shared**, **changes** often, and **matters** when it's wrong.

## Tables, rows and columns

A *relational database* stores data in *tables*. A table is like a spreadsheet with strict rules:

```text
deploys
id | service  | version | environment | status    | started_at
---+----------+---------+-------------+-----------+--------------------
 1 | api      | v1.4.0  | staging     | succeeded | 2026-03-02 09:12:00
 4 | checkout | v2.7.1  | production  | failed    | 2026-03-02 13:02:00
```

- A *column* has a name and a *type*: `INTEGER`, `TEXT`, `REAL`, timestamps and so on. Postgres and MySQL refuse a value of the wrong type; SQLite is looser and stores what it's given unless the table is declared `STRICT`.
- A *row* is one record: one deploy.
- The *schema* is the list of tables and their columns, types and rules. You write it with `CREATE TABLE`.

A table has no built-in order. It's a *set* of rows, not a list, so if you want "newest first" you must ask for it every time.

*SQL* (Structured Query Language) is how you talk to nearly every relational database: SQLite, PostgreSQL, MySQL, SQL Server. The core is the same everywhere; each product adds its own extras.

## Keys: how rows point at each other

### Primary keys

A *primary key* is a column (or a few) that identifies each row: unique, never empty, and stable for the row's whole life. Usually it's an `id` number the database hands out. Names are a bad primary key, because names change: when `checkout` is renamed `cart`, every reference to it would have to change too.

### Foreign keys

A *foreign key* is a column holding another table's primary key. It's how tables point at each other:

```text
teams                     services                         deploys
id | name                 id | name     | team_id          id | service_id | version
---+---------             ---+----------+--------          ---+------------+--------
 1 | payments  ◀───────    2 | checkout |       1  ◀─────   4 |          2 | v2.7.1
 2 | platform  ◀───────    1 | api      |       2  ◀─────   1 |          1 | v1.4.0
```

Deploy 4 has `service_id` 2, so it's a `checkout` deploy; `checkout` has `team_id` 1, so `payments` owns it. With the foreign key declared (`REFERENCES teams (id)`), the database refuses a service whose team doesn't exist, and refuses to delete a team that still owns services. The data can't point at nothing.

### Other constraints

| Constraint | Rule it enforces |
|---|---|
| `NOT NULL` | a value is required |
| `UNIQUE` | no two rows share this value (or combination of values) |
| `CHECK (status IN ('running', 'failed', 'succeeded'))` | only these values |
| `DEFAULT 'standard'` | the value when an insert leaves the column out |

Constraints are your last line of defence: every program that writes to the table, including the one written in a hurry during an incident, has to obey them.

## Asking questions: SELECT

### The shape of a query

```sql
SELECT service, version          -- which columns
FROM deploys                     -- which table
WHERE status = 'failed'          -- which rows
  AND environment = 'production'
ORDER BY started_at DESC         -- in what order
LIMIT 5;                         -- how many
```

You describe **what** you want, not how to find it. The database's *query planner* decides how, which is why the same query can be fast one day and slow the next (see *Indexes* below).

### The order a query really runs in

You write `SELECT` first, but the database works in this order:

```text
FROM      take the table (and JOIN others)
WHERE     keep some rows
GROUP BY  put rows into groups
HAVING    keep some groups
SELECT    compute the output columns
ORDER BY  sort
LIMIT     cut
```

Most confusing errors make sense once you see this. `WHERE COUNT(*) > 3` fails because groups don't exist yet when `WHERE` runs; that's what `HAVING` is for.

### Aggregates and GROUP BY

An *aggregate* boils many rows down to one value: `COUNT(*)`, `SUM`, `AVG`, `MIN`, `MAX`. `GROUP BY` does it once per group:

```sql
SELECT service, COUNT(*) AS failures
FROM deploys
WHERE status = 'failed'
GROUP BY service
HAVING COUNT(*) >= 2;
```

Trace it: `WHERE` keeps the four failed deploys; `GROUP BY` puts them into one pile per service; `COUNT(*)` counts each pile; `HAVING` keeps only piles of two or more. Result: `checkout | 2`.

## Joins: putting tables back together

Keys split facts across tables. A *join* combines them again by matching rows on a condition.

### Inner join: only matches

```sql
SELECT s.name, o.engineer
FROM services AS s
JOIN oncall AS o ON o.service_id = s.id;
```

```text
services          oncall              JOIN result
api       ───▶    api: priya   ───▶   api      | priya
checkout  ───▶    checkout: marco ─▶  checkout | marco
search    ───▶    (no row)            (dropped)
```

A plain `JOIN` is an *inner join*: a row with no partner on the other side disappears.

### Left join: keep everything on the left

```sql
SELECT s.name, o.engineer
FROM services AS s
LEFT JOIN oncall AS o ON o.service_id = s.id;
```

```text
api      | priya
checkout | marco
search   | NULL     ◀── kept, with NULL for the missing side
```

A *left join* keeps every row of the left table and fills the right side with `NULL` when there's no match. Add `WHERE o.service_id IS NULL` and you get exactly the services **nobody** is on call for. Finding what's missing is one of the most useful things you'll do with SQL in ops.

| | inner `JOIN` | `LEFT JOIN` |
|---|---|---|
| left row with a match | kept | kept |
| left row without a match | dropped | kept, right side `NULL` |

## NULL: the value that isn't there

`NULL` means "no value" or "unknown": a deploy that hasn't finished has no `finished_at`. It is not zero and not an empty string, and it follows *three-valued logic*: a comparison can be true, false or **unknown**.

| Expression | Result |
|---|---|
| `NULL = 'production'` | unknown |
| `NULL != 'production'` | unknown |
| `NULL = NULL` | unknown |
| `NULL IS NULL` | true |
| `'v1' \|\| NULL` | `NULL` |
| `COUNT(*)` over 3 rows, one `NULL` | 3 |
| `COUNT(col)` over the same rows | 2 |
| `AVG(col)` | the average of the 2 known values |

`WHERE` keeps only rows where the condition is true, so rows with `NULL` vanish from **both** `= x` and `!= x`. Test with `IS NULL` / `IS NOT NULL`, and fill in defaults with `COALESCE(col, 'unknown')`.

## Indexes

### The full table scan

With no help, `WHERE service = 'api'` must read every row and check it. That's O(n): ten times the rows, ten times the work. With 40 million deploys and a dashboard refreshing every five seconds, the database spends its life reading the same rows.

### A sorted copy with pointers

An *index* is a separate structure that keeps a column's values **sorted**, each pointing at its row. It's usually a *B-tree*: a wide, shallow tree where each step down narrows the search, like the binary search in the *Algorithms and Big-O* guide. Finding `'api'` takes a handful of steps whether the table has a thousand rows or a billion: O(log n).

```sql
CREATE INDEX deploys_by_service ON deploys (service);
```

### Indexes aren't free

| You gain | You pay |
|---|---|
| fast `WHERE`, `JOIN` and `ORDER BY` on the indexed columns | disk space for each index |
| uniqueness checks (`UNIQUE` is an index) | every `INSERT`, `UPDATE`, `DELETE` must update every index |

So index what your queries filter, join and sort on, not everything. A multi-column index `(service, started_at)` is sorted like a phone book (last name, then first name): it helps queries on `service`, or on `service` and `started_at`, but not on `started_at` alone.

### Ask the planner

Don't guess: ask. `EXPLAIN QUERY PLAN` (SQLite) or `EXPLAIN` (Postgres, MySQL) shows how the database will run a query. `SCAN` means "read everything"; `SEARCH … USING INDEX` means it jumps straight to the rows. Wrapping a column in a function, like `WHERE date(started_at) = '2026-03-04'`, usually hides it from its index; compare the bare column with a range instead.

## Transactions and ACID

### All or nothing

Releasing a version is two writes: record the deploy, and update the service's current version. If the database crashes between them, the two tables disagree about what production runs. A *transaction* groups statements so they take effect **together or not at all**:

```sql
BEGIN;
INSERT INTO deploys (service_id, version, status) VALUES (2, 'v2.8.1', 'succeeded');
UPDATE services SET current_version = 'v2.8.1' WHERE id = 2;
COMMIT;   -- or ROLLBACK; to undo everything since BEGIN
```

From Go, `db.Begin()` returns a transaction; you run statements with `tx.Exec` and finish with `tx.Commit()` or `tx.Rollback()`.

### What ACID promises

| Letter | Promise | Without it |
|---|---|---|
| **A**tomicity | all of a transaction's changes happen, or none | half a release recorded |
| **C**onsistency | every constraint holds after each transaction | a deploy pointing at a deleted service |
| **I**solation | concurrent transactions don't see each other's half-done work | a report counting a deploy that's about to be rolled back |
| **D**urability | once `COMMIT` returns, the change survives a crash or power cut | an "approved" change that vanishes after a reboot |

Durability is why databases write each change to a log on disk before they answer "committed". Isolation comes in levels; the defaults of Postgres and MySQL are a sensible middle ground.

## Normalization in plain words

*Normalization* means: **store each fact once**. Look at this table:

```text
deploys
id | service  | team     | slack_channel     | version
---+----------+----------+-------------------+--------
 3 | checkout | payments | #payments-oncall  | v2.7.1
 4 | checkout | payments | #payments-oncall  | v2.7.1
 8 | checkout | payments | #payments-oncal   | v2.7.2   ◀── typo
```

"checkout belongs to payments" and "payments uses #payments-oncall" are repeated on every deploy. When payments changes channel, you must update thousands of rows, and miss some. One row already disagrees. These are *update anomalies*.

The fix is what the keys section showed: `teams`, `services` and `deploys` as separate tables, each fact in one row, connected by foreign keys. Joins reassemble the view you want.

Textbooks call the stages of this *normal forms*. The plain-words test is enough most days: if changing one real-world fact means updating many rows, split the table. *Denormalization* copies data on purpose for speed (say, a reporting table rebuilt nightly); that's fine when it's deliberate and something keeps the copies in step.

## SQL injection and placeholders

### Data that becomes code

Never build a query by gluing in input:

```go
query := "SELECT version FROM deploys WHERE service = '" + service + "'"
```

If `service` is `x' OR '1'='1`, the quote ends your string and the rest becomes SQL: `WHERE service = 'x' OR '1'='1'` matches every row. With `'; DELETE FROM deploys; --`, the input adds its own statement. That's *SQL injection*: input changing the meaning of a query. It's one case of a general problem, data crossing into code, which you'll meet again in the *Security basics for automation* guide with *CLI tools and the OS*.

### Placeholders keep them apart

```go
rows, err := db.Query(
	`SELECT version FROM deploys WHERE service = ?`,
	service,
)
```

The SQL text is fixed, and `service` is sent separately as a *parameter* (a *parameterized query*). The database never parses it as SQL, so whatever it contains, it's only ever compared with `service`. SQLite and MySQL write placeholders as `?`; Postgres as `$1`, `$2`.

Rules that hold everywhere: every value from outside (web forms, URLs, CLI flags, Slack commands, file contents) goes in a placeholder; numbers too; and table or column names never come from input at all (pick them from a fixed list in code).

## SQLite, Postgres and MySQL

| | SQLite | PostgreSQL | MySQL |
|---|---|---|---|
| runs as | a library inside your program; the database is one file | a separate server, reached over the network | a separate server, reached over the network |
| setup | none | install or use a managed service | install or use a managed service |
| many writers at once | one writer at a time | many | many |
| typical use | CLIs, tests, edge devices, small internal tools, local caches | the default choice for new services | huge existing installs, many web apps |
| in the cloud | (ships with your app) | AWS RDS/Aurora, Cloud SQL, Azure | AWS RDS/Aurora, Cloud SQL, Azure |

These exercises use SQLite because a test can create a fresh one in memory in a millisecond. Services shared by many instances usually use Postgres or MySQL, often managed by the cloud provider, which handles backups, failover and patching. The SQL carries over; the details differ (placeholder style, date functions, how strictly types are checked).

## Migrations: changing a schema that's in use

The schema changes as the product does: a new column, a new index, a renamed table. You can't drop and recreate a production table, so changes are written as *migrations*: small, numbered SQL scripts kept in Git next to the code.

```text
migrations/
  0001_create_deploys.sql
  0002_add_services.sql
  0003_add_slo_target.sql      ALTER TABLE services ADD COLUMN slo_target REAL NOT NULL DEFAULT 99.9;
```

A migration tool (golang-migrate, goose, Flyway, Liquibase) records in a table which migrations have run, and applies the new ones in order, usually as a pipeline step before the new code rolls out. Habits that prevent outages:

- **Never edit a migration that has already run** somewhere. Write a new one.
- **Make changes backwards compatible.** During a rolling deploy, old and new code run side by side against the same schema. Add a column (nullable or with a default) first, deploy code that uses it, and only later drop what the old code needed. This is called *expand and contract*.
- **Watch for locks.** Some changes rewrite or lock a big table for minutes. Check your database's docs, try it on a copy of production-sized data, and run risky ones outside peak hours.
- **Run each migration in a transaction** where the database allows it, so a failure leaves the schema as it was.

## Common mistakes

- **`UPDATE` or `DELETE` without `WHERE`.** It changes every row. Run the `WHERE` as a `SELECT` first, and use a transaction so you can `ROLLBACK`.
- **`= NULL` instead of `IS NULL`.** It's never true, so the query quietly returns nothing.
- **Forgetting `NULL` in "not equal" filters.** `status != 'failed'` also drops the rows where status is `NULL`.
- **Filtering a `LEFT JOIN` in `WHERE`.** A condition on the right table in `WHERE` throws away the unmatched rows and turns it back into an inner join; put it in `ON`.
- **`COUNT(*)` after a `LEFT JOIN`.** It counts the empty match as 1; count a column from the right table.
- **Building SQL with `+` or `fmt.Sprintf`.** That's SQL injection waiting to happen. Use placeholders for every value.
- **Relying on row order without `ORDER BY`.** It works until the data grows or an index appears.
- **Indexing everything, or nothing.** Index the columns real queries filter, join and sort on, then check with `EXPLAIN`.
- **Two related writes outside a transaction.** A crash between them leaves the data inconsistent.

## Check yourself

1. Two CI pipelines append a line to the same `deploys.log` at the same moment, and a third crashes mid-write. Name two things a database would do better here.
2. Why is a numeric `id` a better primary key for a `services` table than the service's name?
3. A `deploys` table has a `service_id` column declared `REFERENCES services (id)`. What does the database now refuse to do?
4. Why does `SELECT service, COUNT(*) FROM deploys WHERE COUNT(*) > 3 GROUP BY service` fail, and how do you fix it?
5. `SELECT * FROM deploys WHERE finished_at = NULL` returns no rows, although some deploys are still running. Why, and what should it say?
6. You need every service with its number of deploys, including services that never deployed (shown as 0). Which join and which count do you use?
7. A query `WHERE service = 'api'` on a 40-million-row table takes seconds. What does an index on `service` change, in Big-O terms, and what does it cost?
8. A release inserts a deploy row and then updates the service's current version. What can go wrong without a transaction, and which letter of ACID covers it?
9. A Go function runs `db.Query("SELECT * FROM deploys WHERE service = '" + name + "'")` with `name` taken from a URL. What can an attacker do, and what is the fix?
10. Why do teams write schema changes as numbered migration files instead of editing the schema by hand on the server?

## Answers

1. A database coordinates concurrent writers so their records don't interleave, and makes each write atomic, so a crash leaves either the whole record or nothing. It can also enforce rules on values and answer new questions with a query instead of a new script.
2. Names change (checkout becomes cart), and every row pointing at the old name would have to change too. An `id` never changes and is unique, so references stay valid; the name is just a column you can update in one place.
3. It refuses a deploy whose `service_id` matches no service, and refuses to delete (or change the `id` of) a service that still has deploys. Rows can't point at nothing. (In SQLite this needs `PRAGMA foreign_keys = ON`.)
4. `WHERE` runs on single rows before groups exist, so it can't use an aggregate. Move the condition after `GROUP BY service` as `HAVING COUNT(*) > 3`.
5. Any comparison with `NULL` using `=` is unknown, not true, and `WHERE` only keeps true rows. Write `WHERE finished_at IS NULL`.
6. `services LEFT JOIN deploys ON deploys.service_id = services.id`, grouped by the service, with `COUNT(deploys.id)`. The left join keeps services without deploys; `COUNT(*)` would count their empty match as 1, while `COUNT(deploys.id)` skips the `NULL` and gives 0.
7. Without an index the database scans every row: O(n). A B-tree index lets it jump to the matching rows in O(log n) steps. The cost is disk space and slower writes, since every insert, update and delete must also update the index.
8. If the database or program fails between the two statements, the deploy is recorded but the service still shows the old version, so the tables disagree. A transaction makes both happen or neither: atomicity, the A in ACID.
9. The input can close the quoted string and add its own SQL: `x' OR '1'='1` returns every row, and `'; DELETE FROM deploys; --` could delete data. Use a placeholder: `db.Query("SELECT * FROM deploys WHERE service = ?", name)`, so the value is sent separately and never parsed as SQL.
10. Migration files live in Git, so changes are reviewed, versioned and applied in the same order to every environment (dev, staging, production) by a tool that records which ones already ran. Hand edits are invisible, unrepeatable and easy to get wrong on one server.

## Key terms

| Term | Meaning |
|---|---|
| database | a program that stores data for other programs and answers queries about it |
| relational database | a database that keeps data in tables of rows and columns linked by keys |
| SQL | Structured Query Language: the language for querying and changing relational databases |
| table | a named set of rows that all have the same columns |
| schema | the tables, columns, types and rules of a database |
| primary key | the column(s) that uniquely identify each row of a table |
| foreign key | a column holding another table's primary key, so rows can point at each other |
| constraint | a rule the database enforces on every write, such as NOT NULL, UNIQUE or CHECK |
| query planner | the part of the database that decides how to run a query |
| aggregate | a function that turns many rows into one value, such as COUNT, SUM or AVG |
| inner join | combines matching rows of two tables and drops rows without a match |
| left join | keeps every row of the left table, with NULLs where the right table has no match |
| NULL | the marker for a missing or unknown value; comparisons with it are unknown |
| index | a sorted structure over some columns that lets the database find rows without reading them all |
| full table scan | reading every row of a table to answer a query: O(n) |
| transaction | a group of statements that take effect together or not at all |
| ACID | atomicity, consistency, isolation, durability: the guarantees of a transaction |
| normalization | designing tables so each fact is stored once |
| SQL injection | input that changes the meaning of a query because it was glued into the SQL text |
| parameterized query | a query with placeholders whose values are sent separately from the SQL |
| migration | a versioned script that changes a database's schema in place |
| three-valued logic | SQL's logic where a condition is true, false or unknown, because comparing with NULL is unknown |
| B-tree | the wide, shallow, sorted tree most database indexes use; a lookup takes O(log n) steps |
| update anomaly | data disagreeing with itself because one fact was stored in many rows and only some were changed |
| denormalization | copying data on purpose for speed, accepting the work of keeping the copies in step |
| expand and contract | changing a schema in compatible steps: add the new thing, move code over, then remove the old thing |

## Watch

- [SQL Explained in 100 Seconds](https://www.youtube.com/watch?v=zsjvFFKOm3c) · Fireship · 2 min: the whole landscape (tables, keys, joins, the big databases) in one quick pass; watch it first.
- [Database Indexing Explained (with PostgreSQL)](https://www.youtube.com/watch?v=-qNSXK7s7_w) · Hussein Nasser · 18 min: runs queries on a big Postgres table with and without an index and reads the query plans: the *Indexes* section made real.
- [ACID Properties in Databases With Examples](https://www.youtube.com/watch?v=GAe5oB742dw) · ByteByteGo · 5 min: the four letters with clear diagrams, good for revising the transactions section.
- [Learn Database Normalization - 1NF, 2NF, 3NF, 4NF, 5NF](https://www.youtube.com/watch?v=GFQaEYEc8_8) · Decomplexify · 29 min: a calm, example-driven walk through the normal forms; the first half, up to third normal form, covers everything you need now.
- [Running an SQL Injection Attack - Computerphile](https://www.youtube.com/watch?v=ciNHn38EyRc) · Computerphile · 17 min: Mike Pound attacks a deliberately vulnerable test site live, so you see what glued-together SQL lets an attacker do.

## Go deeper

- [Go: accessing relational databases](https://go.dev/doc/database/): the official guide to `database/sql`: opening, querying, transactions and avoiding injection.
- [Appropriate Uses For SQLite](https://www.sqlite.org/whentouse.html): SQLite's own honest account of when it fits and when you want a client/server database instead.
- [Use The Index, Luke](https://use-the-index-luke.com/): a free, readable book on how indexes work and how to write queries that use them.
- [PostgreSQL tutorial](https://www.postgresql.org/docs/current/tutorial.html): the official tutorial, for when you move from SQLite to the database most services run on.
- [Database normalization on Wikipedia](https://en.wikipedia.org/wiki/Database_normalization): the formal normal forms, with examples, once the plain-words version feels comfortable.
