## Task 1: guard first

Start `Rollback` with an early return when `tx.done` is true. After appending
`rollback`, set `tx.done = true` so a second call does nothing.

## Task 2: the transaction pattern

Follow the pattern from the Learn section: `tx := db.Begin()`, then
`defer tx.Rollback()` on the next line. Loop with `for i, change := range
changes` so you have the index for the `change <n>: ` prefix (`i+1`).

## Task 3: collect, then join

Make a `var errs []error`. In the loop, call `ApplyChanges` and append
`fmt.Errorf("batch %d: %w", i+1, err)` when it fails. Finish with
`return errors.Join(errs...)`, which is `nil` when `errs` is empty.

## Almost the answer

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
