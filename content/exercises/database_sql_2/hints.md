## Task 1: Stop the search box injection

The query is built by gluing `service` into the text, so a `'` in the input
ends your string early. Make the SQL text fixed and send `service`
separately.

### Closer

Replace `'" + service + "'` (both quotes included) with a single `?`, and pass
`service` as the second argument to `db.Query`. You no longer need the
`query` variable.

### Answer

```go
rows, err := db.Query(
	`SELECT version FROM deploys WHERE service = ? AND status = 'failed' ORDER BY id`,
	service,
)
```

The rest of the function stays the same.

## Task 2: Stop the rollback injection

Same bug as task 1, with a number. `id` is text from a URL, so it can carry
`OR 1=1` along with it.

### Closer

End the SQL with `WHERE id = ?` and pass `id` after it:
`db.Exec("UPDATE deploys SET status = 'rolled_back' WHERE id = ?", id)`.

### Answer

```go
res, err := db.Exec(`UPDATE deploys SET status = 'rolled_back' WHERE id = ?`, id)
```

SQLite turns `"14"` into the number 14 for the comparison, but
`"14 OR 1=1"` matches no id, so `RowsAffected` is 0 and you return
`ErrNotFound`.

## Task 3: Read running deploys

`finished_at` can be `NULL`, and a `string` has no way to say "no value".
Scan that column into a type that can, then copy its text into the struct.

### Closer

Declare `var finished sql.NullString`, scan `&finished` instead of
`&d.FinishedAt`, and after the error checks set `d.FinishedAt =
finished.String`.

### Answer

```go
var d Deploy
var finished sql.NullString
err := db.QueryRow(
	`SELECT id, service, version, status, finished_at FROM deploys WHERE id = ?`,
	id,
).Scan(&d.ID, &d.Service, &d.Version, &d.Status, &finished)
if errors.Is(err, sql.ErrNoRows) {
	return Deploy{}, ErrNotFound
}
if err != nil {
	return Deploy{}, fmt.Errorf("get deploy %d: %w", id, err)
}
d.FinishedAt = finished.String // "" when finished_at is NULL
return d, nil
```

## Task 4: Make the release all or nothing

Start a transaction with `db.Begin()`, run both statements on the
transaction instead of on `db`, and commit at the end. Arrange for a
rollback on every early return.

### Closer

`tx, err := db.Begin()`, then `defer tx.Rollback()`. Change both `db.Exec` to
`tx.Exec`. Finish with `if err := tx.Commit(); err != nil { ... }`.

### Answer

```go
tx, err := db.Begin()
if err != nil {
	return fmt.Errorf("release %s %s: %w", service, version, err)
}
defer tx.Rollback()

_, err = tx.Exec(
	`INSERT INTO deploys (service, version, environment, status, started_at)
	 VALUES (?, ?, 'production', 'succeeded', ?)`,
	service, version, startedAt,
)
if err != nil {
	return fmt.Errorf("release %s %s: %w", service, version, err)
}
_, err = tx.Exec(`UPDATE services SET current_version = ? WHERE name = ?`, version, service)
if err != nil {
	return fmt.Errorf("release %s %s: %w", service, version, err)
}
if err := tx.Commit(); err != nil {
	return fmt.Errorf("release %s %s: %w", service, version, err)
}
return nil
```

If the `UPDATE` fails, the function returns and the deferred `Rollback`
removes the inserted row.
