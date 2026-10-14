## Task 1: Make Rollback safe to call twice

`tx.done` tells you whether the transaction is already finished. Check it
before doing anything, and remember to update it once you have rolled back.

### Closer

Start with an early return: `if tx.done { return }`. After appending
`rollback`, set `tx.done = true` so a second call does nothing.

### Answer

```go
if tx.done {
	return
}
tx.db.Events = append(tx.db.Events, "rollback")
tx.done = true
```

The guard is what makes `defer tx.Rollback()` harmless after a `Commit`.

## Task 2: Apply changes in one transaction

Follow the pattern from the Learn section: begin, defer the rollback right
away, run each change, and commit at the end. You need the loop index for the
`change <n>: ` prefix.

### Closer

`tx := db.Begin()`, then `defer tx.Rollback()` on the next line. Loop with
`for i, change := range changes` and use `i+1` in
`fmt.Errorf("change %d: %w", i+1, err)`. Finish with `return tx.Commit()`.

### Answer

```go
func ApplyChanges(db *DB, changes []string) error {
	tx := db.Begin()
	defer tx.Rollback()

	for i, change := range changes {
		if err := tx.Exec(change); err != nil {
			return fmt.Errorf("change %d: %w", i+1, err)
		}
	}
	return tx.Commit()
}
```

On an early return the deferred `Rollback` undoes the work; after `Commit` it
does nothing.

## Task 3: Apply every batch in its own transaction

Call `ApplyChanges` once per batch and don't stop on failure. Collect the
errors as you go and combine them at the end.

### Closer

Make a `var errs []error`. When a batch fails, append
`fmt.Errorf("batch %d: %w", i+1, err)`. Finish with
`return errors.Join(errs...)`.

### Answer

```go
var errs []error
for i, batch := range batches {
	if err := ApplyChanges(db, batch); err != nil {
		errs = append(errs, fmt.Errorf("batch %d: %w", i+1, err))
	}
}
return errors.Join(errs...)
```

`errors.Join` returns `nil` when `errs` is empty.
