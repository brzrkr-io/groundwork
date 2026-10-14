## Task 1: Count a service's failed deploys

`SELECT COUNT(*)` always returns exactly one row, so use the method for one
row, then copy the number into an `int`. The service name is a value, so it
travels as an argument, not inside the SQL.

### Closer

`db.QueryRow("SELECT COUNT(*) FROM deploys WHERE service = ? AND status = 'failed'", service).Scan(&n)`.
On error, return `0` and `fmt.Errorf("count failed deploys: %w", err)`.

### Answer

```go
var n int
err := db.QueryRow(
	`SELECT COUNT(*) FROM deploys WHERE service = ? AND status = 'failed'`,
	service,
).Scan(&n)
if err != nil {
	return 0, fmt.Errorf("count failed deploys: %w", err)
}
return n, nil
```

`'failed'` is fixed text, so it may stay in the SQL; only the input needs a
placeholder.

## Task 2: Get one deploy by id

Select the six columns in the order of the struct's fields and `Scan` into
the address of each field. Then handle the "no row" error before any other
error.

### Closer

`.Scan(&d.ID, &d.Service, &d.Version, &d.Environment, &d.Status, &d.StartedAt)`.
Then `if errors.Is(err, sql.ErrNoRows) { return Deploy{}, ErrNotFound }`.

### Answer

```go
var d Deploy
err := db.QueryRow(
	`SELECT id, service, version, environment, status, started_at
	 FROM deploys WHERE id = ?`,
	id,
).Scan(&d.ID, &d.Service, &d.Version, &d.Environment, &d.Status, &d.StartedAt)
if errors.Is(err, sql.ErrNoRows) {
	return Deploy{}, ErrNotFound
}
if err != nil {
	return Deploy{}, fmt.Errorf("get deploy %d: %w", id, err)
}
return d, nil
```

## Task 3: List a service's latest deploys

Same columns as task 2, but many rows: `db.Query`, then a `for rows.Next()`
loop that scans each row into a fresh `Deploy` and appends it. Don't forget
the two things that come after: closing the rows and checking `rows.Err()`.

### Closer

The SQL ends with `ORDER BY started_at DESC LIMIT ?`, and you pass
`service, limit`. Put `defer rows.Close()` right after the `Query` error
check.

### Answer

```go
rows, err := db.Query(
	`SELECT id, service, version, environment, status, started_at
	 FROM deploys
	 WHERE service = ?
	 ORDER BY started_at DESC
	 LIMIT ?`,
	service, limit,
)
if err != nil {
	return nil, fmt.Errorf("recent deploys: %w", err)
}
defer rows.Close()

var deploys []Deploy
for rows.Next() {
	var d Deploy
	if err := rows.Scan(&d.ID, &d.Service, &d.Version, &d.Environment, &d.Status, &d.StartedAt); err != nil {
		return nil, fmt.Errorf("recent deploys: %w", err)
	}
	deploys = append(deploys, d)
}
if err := rows.Err(); err != nil {
	return nil, fmt.Errorf("recent deploys: %w", err)
}
return deploys, nil
```

## Task 4: Record a new deploy

An `INSERT` returns no rows, so use `Exec`. Leave `id` out of the column
list so the database picks it, and pass the five fields as arguments.

### Closer

`res, err := db.Exec("INSERT INTO deploys (service, version, environment, status, started_at) VALUES (?, ?, ?, ?, ?)", d.Service, ...)`,
then `res.LastInsertId()`, which also returns an error.

### Answer

```go
res, err := db.Exec(
	`INSERT INTO deploys (service, version, environment, status, started_at)
	 VALUES (?, ?, ?, ?, ?)`,
	d.Service, d.Version, d.Environment, d.Status, d.StartedAt,
)
if err != nil {
	return 0, fmt.Errorf("record deploy: %w", err)
}
id, err := res.LastInsertId()
if err != nil {
	return 0, fmt.Errorf("record deploy: %w", err)
}
return id, nil
```
