# INSERT, UPDATE and DELETE: record a release

Queries only read. A deploy tool also *writes*: it records each deploy,
updates its status, cleans up old rows, and keeps two tables in step. One
missing `WHERE` here changes every row in the table, so this is where SQL
gets careful.

**New here:** `INSERT INTO … VALUES`, `UPDATE … SET … WHERE`, `DELETE FROM … WHERE`, transactions with `BEGIN` and `COMMIT`.

## Learn

### The tables

Open `testdata/schema.sql` and `testdata/data.sql`: `services` (with the
version production runs now, `current_version`) and 18 `deploys`, each
pointing at a service by `service_id`, as in *Joins: services, teams and who
to page*. Each test loads a fresh copy, runs your file, then checks what the
tables hold.

### INSERT adds rows

```sql
INSERT INTO services (name, current_version)
VALUES ('notifications', 'v0.1.0');
```

List the columns, then the values in the **same order**. A column you leave
out gets its default; an `INTEGER PRIMARY KEY` left out gets the next free
number, so you rarely write the `id` yourself.

### UPDATE changes rows

```sql
UPDATE services
SET current_version = 'v0.1.1'
WHERE name = 'notifications';
```

`SET` lists the new values (`SET a = 1, b = 2` for several). `WHERE` picks
the rows, exactly as in a `SELECT`.

> **Why is it written like this?** SQL makes `WHERE` optional, and an
> `UPDATE` or `DELETE` without one changes **every row**: no warning, no
> undo. A habit that saves careers: write the `WHERE` first, and run it as a
> `SELECT` to see which rows it hits before you change them. Picking a row by
> its `id` is the safest choice, because ids are unique.

### DELETE removes rows

```sql
DELETE FROM deploys
WHERE status = 'failed'
  AND started_at < '2025-06-01';
```

Same rule: the `WHERE` decides which rows go. `DELETE FROM deploys;` empties
the table.

### Transactions: all or nothing

Releasing a version takes two changes: record the deploy, and update the
service's `current_version`. If the second fails (disk full, connection
lost) after the first succeeded, the tables disagree about what's running.

A *transaction* groups statements so they happen **all together or not at
all**. Retiring a service, for example, deletes its deploys and then the
service itself; half of that would leave a mess:

```sql
BEGIN;
DELETE FROM deploys WHERE service_id = 3;
DELETE FROM services WHERE id = 3;
COMMIT;
```

| Statement | Means |
|---|---|
| `BEGIN` | start a transaction |
| `COMMIT` | keep every change since `BEGIN`, all at once |
| `ROLLBACK` | throw away every change since `BEGIN` |

Until `COMMIT`, other connections don't see the changes. If a statement
fails, it usually undoes only itself; the transaction stays open, and the
program sends `ROLLBACK` to put the database back exactly as it was. Without
`BEGIN`, each statement is its own tiny transaction, committed the moment it
finishes, so nothing can undo it later.

## Your task

You're writing the SQL behind a deploy tool's buttons. Each file is run once
against a fresh copy of the tables.

### 1. Record a new deploy

**Goal:** `1_record_deploy.sql` adds one deploy: `service_id` 4 (search),
`version` `v3.2.0`, `environment` `staging`, `status` `running`,
`started_at` `2026-03-07 09:00:00`. Let the database pick the `id`.

Afterwards there are 19 deploys, and the new one has id 19.

**Think it through:**
- Which five columns do you list, and in which order do the values follow?
- Which values need quotes?

### 2. Mark one deploy as rolled back

**Goal:** `2_mark_rolled_back.sql` sets the `status` of deploy **14** to
`rolled_back`, and changes nothing else.

The test compares every row before and after; any other row that changed
is listed.

**Think it through:**
- Which column identifies exactly one deploy?
- What happens to the other 17 rows if the `WHERE` is missing?

### 3. Delete last year's staging deploys

**Goal:** `3_delete_old_staging.sql` deletes every **staging** deploy that
started **before 2026**, and keeps every production deploy, however old.

Afterwards the ids left are 1 to 14, 17 and 18.

**Think it through:**
- Two conditions. How do you write "before 2026" for a text timestamp?
- Try your `WHERE` in a `SELECT` first: does it pick only 15 and 16?

### 4. Release checkout in one transaction

**Goal:** `4_release.sql` does two things inside one transaction (`BEGIN;`
… `COMMIT;`), in this order:

1. insert a deploy: `service_id` 2 (checkout), `version` `v2.8.1`,
   `environment` `production`, `status` `succeeded`, `started_at`
   `2026-03-07 10:00:00`;
2. set checkout's (`id` 2) `current_version` to `v2.8.1`.

The test also runs your file against a database where updating `services`
fails, and checks that the new deploy row disappears with it.

**Think it through:**
- Where do `BEGIN` and `COMMIT` go relative to the two changes?
- Without them, is the inserted row still there after the update fails?

## Learn more

- [SQLite: INSERT](https://www.sqlite.org/lang_insert.html)
- [SQLite: UPDATE](https://www.sqlite.org/lang_update.html)
- [SQLite: transactions](https://www.sqlite.org/lang_transaction.html)
