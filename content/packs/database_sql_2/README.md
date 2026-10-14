# database/sql: fix the SQL injection

The deploy status API works fine until someone types `x' OR '1'='1` into
the search box, or calls `/deploys/14 OR 1=1/rollback`. Then it leaks data
and rolls back every deploy you have. Fix that, plus two more classic
`database/sql` bugs: scanning a `NULL`, and two writes that should have been
one transaction.

**New here:** SQL injection, `sql.NullString`, `db.Begin`, `tx.Exec`, `tx.Commit`, `tx.Rollback`, `RowsAffected`.

## Learn

### Recap

As in *database/sql: query deploys from Go*: `db.Query`, `db.QueryRow` and
`db.Exec` take the SQL text, then the values for its `?` placeholders.
`Scan` copies columns into pointers. The test builds the database from
`testdata/schema.sql` and `testdata/data.sql`; deploy 15 is still running.

### SQL injection

Look at how the starter builds a query:

```go
query := "SELECT version FROM deploys WHERE service = '" + service + "'"
```

With `service` = `checkout`, the database receives what you meant. With
`service` = `x' OR '1'='1`, it receives:

```sql
SELECT version FROM deploys WHERE service = 'x' OR '1'='1'
```

The quote in the input **closed your string**, and the rest became SQL:
`'1'='1'` is true for every row. With `'; DELETE FROM deploys; --` the input
adds a whole new statement, and `--` comments out what's left of yours.
That's *SQL injection*: input that changes the meaning of a query. It has
been near the top of the OWASP list of web security risks for twenty years.

Escaping quotes by hand is a losing game. The fix is a placeholder:

```go
db.Query(`SELECT version FROM deploys WHERE service = ?`, service)
```

The SQL text is fixed, and `service` travels separately as a value. Whatever
it contains, it can only ever be compared with `service`.

> **Why is it written like this?** No compiler or vet check stops you from
> concatenating SQL: the rule "values always go in placeholders" is yours to
> keep. It applies to numbers too: an `id` that came from a URL is text an
> attacker controls. Placeholders only work for *values*, not for table or
> column names; never build those from input.

### RowsAffected

`res.RowsAffected()` from `Exec` says how many rows changed. A `0` after an
`UPDATE ... WHERE id = ?` means no row had that id, so the starter's
`MarkRolledBack` returns `ErrNotFound`. Keep that part.

### Scanning NULL

`Scan` into a `string` fails if the column is `NULL`:
`converting NULL to string is unsupported`. For a column that may be `NULL`,
scan into a `sql.NullString`:

```go
var owner sql.NullString
err := db.QueryRow(`SELECT owner FROM services WHERE name = ?`, name).Scan(&owner)
// owner.Valid is false for NULL; owner.String holds the text, "" for NULL
```

`sql.NullInt64`, `sql.NullTime` and friends work the same way.

### Transactions from Go

As in *INSERT, UPDATE and DELETE: record a release*, two writes that must
happen together belong in a transaction. In Go:

```go
tx, err := db.Begin()
if err != nil {
	return err
}
defer tx.Rollback() // undoes everything unless Commit ran first

if _, err := tx.Exec(`UPDATE services SET current_version = ? WHERE name = ?`, v, name); err != nil {
	return err // the deferred Rollback undoes the changes
}
if _, err := tx.Exec(`INSERT INTO audit (message) VALUES (?)`, "released "+v); err != nil {
	return err
}
return tx.Commit()
```

| Call | Does |
|---|---|
| `db.Begin()` | starts a transaction on one connection, returns a `*sql.Tx` |
| `tx.Exec`, `tx.Query`, `tx.QueryRow` | run statements **inside** it |
| `tx.Commit()` | keeps the changes |
| `tx.Rollback()` | throws them away; after a `Commit` it does nothing |

> **Why is it written like this?** `defer tx.Rollback()` right after `Begin`
> is the common Go pattern: every early `return` is then safe, and you can't
> forget the cleanup (see *defer*). Inside the transaction, use `tx.Exec`,
> not `db.Exec`: `db` runs on another connection, outside your transaction,
> and here it fails with `database is locked`.

## Your task

Fix the four functions in `exercise.go`. Keep their error wrapping as it is.

### 1. Stop the search box injection

**Goal:** `FailedVersions(db, service)` passes `service` as a `?`
placeholder value, so no input can change the query.

```go
FailedVersions(db, "checkout")                   // => ["v2.7.1" "v2.8.0"], nil
FailedVersions(db, "x' OR '1'='1")               // => [], nil (no such service)
FailedVersions(db, "o'brien-api")                // => [], nil, not a syntax error
FailedVersions(db, "'; DELETE FROM deploys; --") // => [], nil, and all 15 deploys still there
```

**Think it through:**
- Which part of the query text is the value? Replace it, quotes included,
  with `?`.
- Where does `service` go instead?

### 2. Stop the rollback injection

**Goal:** `MarkRolledBack(db, id)` passes `id` as a placeholder value.

```go
MarkRolledBack(db, "14")        // => nil: deploy 14 is now rolled_back
MarkRolledBack(db, "999")       // => ErrNotFound
MarkRolledBack(db, "14 OR 1=1") // => ErrNotFound, and no deploy changed
```

**Think it through:**
- `"14 OR 1=1"` passed as a value is compared with each `id` as a whole.
  Does any id equal it?

### 3. Read running deploys

**Goal:** `GetDeploy(db, id)` works for deploys whose `finished_at` is
`NULL`, setting `FinishedAt` to `""` for them.

```go
GetDeploy(db, 4)  // => Deploy{ID: 4, ..., Status: "failed", FinishedAt: "2026-03-02 13:12:10"}, nil
GetDeploy(db, 15) // => Deploy{ID: 15, ..., Status: "running", FinishedAt: ""}, nil
```

**Think it through:**
- Which `Scan` target can hold a `NULL`?
- After a successful `Scan`, which of its fields goes into `d.FinishedAt`?

### 4. Make the release all or nothing

**Goal:** `Release(db, service, version, startedAt)` runs its `INSERT` and its
`UPDATE` in one transaction, so a failing `UPDATE` leaves no new deploy row.

The test makes the `UPDATE` fail on purpose and checks that the inserted row
is gone. On success, both changes must be saved. Either way, the transaction
must be over when `Release` returns: an open one keeps its connection busy
and blocks other writers.

**Think it through:**
- What do you call to start the transaction, and what must each statement
  run on instead of `db`?
- What undoes the insert on every error path, and what saves both changes at
  the end?

## Learn more

- [Go: avoiding SQL injection risk](https://go.dev/doc/database/sql-injection)
- [Go: executing transactions](https://go.dev/doc/database/execute-transactions)
- [OWASP: SQL injection prevention cheat sheet](https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html)
