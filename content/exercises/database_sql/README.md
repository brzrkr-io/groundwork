# database/sql: query deploys from Go

Your deploy tool, status page and Slack bot all need the deploy history, so
they talk to the database from Go. Go's standard `database/sql` package does
that the same way for SQLite, Postgres and MySQL. Here you write the four
functions every data-access layer has: count, get one, list, insert.

**New here:** `database/sql`, drivers, `*sql.DB`, `QueryRow`, `Query`, `Exec`, `Scan`, `rows.Next`, `rows.Err`, `sql.ErrNoRows`, `?` placeholders, `LastInsertId`.

## Learn

### database/sql and drivers

`database/sql` defines **one** API for every SQL database. A *driver*
package teaches it a particular database and registers itself under a name
when imported:

```go
import (
	"database/sql"

	_ "modernc.org/sqlite" // registers the driver "sqlite"
)

db, err := sql.Open("sqlite", "deploys.db")
```

Your functions take a `*sql.DB` as a parameter, so they don't care which
database is behind it. The test opens a SQLite file in a temporary folder,
loads `testdata/schema.sql` and `testdata/data.sql` (the familiar `deploys`
table), and passes it in. That's why only the test file imports the driver.

> **Why is it written like this?** The `_` import is required: you never call
> the driver package by name, and Go rejects unused imports, so `_` says
> "import it only for its side effect" (registering itself). A `*sql.DB` is a
> pool of connections that's safe to share; by convention a program opens it
> once at start-up and passes it around, as here.

### Placeholders: values travel separately

Never build a query by gluing strings. Put a `?` where each value goes and
pass the values after the query:

```go
db.QueryRow(`SELECT COUNT(*) FROM deploys WHERE environment = ?`, env)
```

The driver sends the SQL and the values **separately**, so a value is always
just data, even if it contains a `'`. Gluing it in with `+` or `fmt.Sprintf`
lets the input rewrite your query: that's *SQL injection*, and
*database/sql: fix the SQL injection* is all about it. Postgres writes
placeholders as `$1`, `$2`; SQLite and MySQL use `?`.

### One row: QueryRow and Scan

`QueryRow` runs a query you expect to return one row. `Scan` copies its
columns, in order, into the variables you point it at:

```go
var service, status string
err := db.QueryRow(
	`SELECT service, status FROM deploys WHERE id = ?`, 7,
).Scan(&service, &status)
```

If there was no row, `Scan` returns `sql.ErrNoRows`. Check for it with
`errors.Is` and turn it into your own error, so callers don't depend on
`database/sql` details (see *Errors are values* and *Wrapping errors*).

> **Why is it written like this?** `Scan` needs pointers (`&service`) because
> it must change your variables, like `json.Unmarshal` in *Decoding JSON:
> load a service config*. The number and order of pointers must match the
> `SELECT` list exactly. Backquotes (raw strings) are a convention for SQL in
> Go: they can span lines and hold `'` without escaping.

### Many rows: Query, Next, Scan, Err

```go
rows, err := db.Query(`SELECT version FROM deploys WHERE service = ?`, svc)
if err != nil {
	return nil, err
}
defer rows.Close()

var versions []string
for rows.Next() {
	var v string
	if err := rows.Scan(&v); err != nil {
		return nil, err
	}
	versions = append(versions, v)
}
if err := rows.Err(); err != nil {
	return nil, err
}
return versions, nil
```

1. `Query` starts the query. On success you **must** `Close` the rows, or the
   connection is never returned to the pool: `defer rows.Close()` (see
   *defer*).
2. `rows.Next()` moves to the next row, returning `false` at the end.
3. `rows.Scan` reads the current row.
4. `rows.Err()` reports an error that ended the loop early. Skip it and a
   network failure looks like "fewer rows".

`LIMIT ?` takes a placeholder too.

### Changing data: Exec

`Exec` runs a statement that returns no rows (`INSERT`, `UPDATE`,
`DELETE`). Its `sql.Result` says what happened:

```go
res, err := db.Exec(`UPDATE deploys SET status = ? WHERE id = ?`, "failed", 3)
n, err := res.RowsAffected()  // how many rows changed
id, err := res.LastInsertId() // after an INSERT: the id the database chose
// (check each err as usual; left out here to keep the example short)
```

`LastInsertId` works with SQLite and MySQL. Postgres drivers don't support
it: there you add `RETURNING id` to the `INSERT` and read the id with
`QueryRow(...).Scan(&id)`.

## Your task

Write the data-access functions for the deploy tool in `exercise.go`. Wrap
errors with `fmt.Errorf` and `%w` using the prefix each TODO gives; add
`"fmt"` to the imports.

### 1. Count a service's failed deploys

**Goal:** `CountFailed(db, service)` returns how many deploys of `service`
have the status `failed`.

```go
CountFailed(db, "checkout")        // => 2, nil
CountFailed(db, "api")             // => 0, nil
CountFailed(db, "checkout' OR '1'='1") // => 0, nil: it's just a strange name
```

**Think it through:**
- One number comes back. Which method returns one row?
- Where does `service` go: inside the SQL text, or after it?

### 2. Get one deploy by id

**Goal:** `GetDeploy(db, id)` returns the deploy with that `id`, with all six
fields filled, or `Deploy{}` and `ErrNotFound` if there is none.

```go
GetDeploy(db, 4)
// => Deploy{ID: 4, Service: "checkout", Version: "v2.7.1",
//    Environment: "production", Status: "failed", StartedAt: "2026-03-02 13:02:00"}, nil
GetDeploy(db, 999) // => Deploy{}, ErrNotFound
```

**Think it through:**
- List the columns in the same order as the fields you `Scan` into.
- Which error means "no row", and how do you compare errors safely?

### 3. List a service's latest deploys

**Goal:** `RecentDeploys(db, service, limit)` returns at most `limit` deploys
of `service`, **newest first** (by `started_at`), with all fields filled. A
service with no deploys gives an empty slice and no error.

```go
RecentDeploys(db, "api", 2)     // => deploys 13 and 9
RecentDeploys(db, "billing", 10) // => deploys 11, 10 and 5
```

**Think it through:**
- Two placeholders, two values, in the order the `?`s appear.
- Three places can fail: `Query`, `Scan`, and the check after the loop.

### 4. Record a new deploy

**Goal:** `RecordDeploy(db, d)` inserts `d`'s `Service`, `Version`,
`Environment`, `Status` and `StartedAt` as a new row, **ignoring `d.ID`**,
and returns the id the database gave it.

```go
RecordDeploy(db, Deploy{Service: "search", Version: "v3.2.0",
	Environment: "staging", Status: "running", StartedAt: "2026-03-07 09:00:00"})
// => 15, nil (the table had ids 1 to 14)
```

A service named `o'brien-api` must be stored exactly as written.

**Think it through:**
- Which method runs a statement that returns no rows?
- Which method of its result gives the new id?

## Learn more

- [Go: accessing relational databases](https://go.dev/doc/database/)
- [database/sql package](https://pkg.go.dev/database/sql)
- [Go: querying for data](https://go.dev/doc/database/querying)
