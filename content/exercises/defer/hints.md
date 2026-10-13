## Task 1: Always close the connection in Deploy

The early `return` inside the loop skips the `r.Close(host)` at the end. Use
`defer` so the close runs on every return path.

### Closer

Put `defer r.Close(host)` on the line right after `r.Connect(host)`. Then
delete the old `r.Close(host)` near the end, or the connection is closed twice.

### Answer

```go
r.Connect(host)
defer r.Close(host)
for _, cmd := range cmds {
	if err := r.Run(cmd); err != nil {
		return err
	}
}
return nil
```

## Task 2: Lock, connect and clean up in order in Migrate

Each thing you acquire gets its own `defer` straight after it. Write them in
the order you acquire things; last-in, first-out handles the closing order.

### Closer

Lock, then `defer r.Unlock("migrations")`. Connect, then `defer r.Close(host)`.
Finish with `return r.Run("migrate")`.

### Answer

```go
func Migrate(r *Recorder, host string) error {
	r.Lock("migrations")
	defer r.Unlock("migrations")
	r.Connect(host)
	defer r.Close(host)
	return r.Run("migrate")
}
```

The close was deferred last, so it runs first, before the unlock.

## Task 3: Add context to every Backup error

The result is named `err`, so a deferred closure can read and change it after
a `return` has stored the error there.

### Closer

```go
defer func() {
	if err != nil {
		// replace err with a wrapped version
	}
}()
```

### Answer

```go
defer func() {
	if err != nil {
		err = fmt.Errorf("backup %s: %w", host, err)
	}
}()
```

Checking for `nil` first keeps a successful backup returning `nil`.
