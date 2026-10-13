## Task 1: one defer, one deleted line

Put `defer r.Close(host)` on the line right after `r.Connect(host)`. Then
delete the old `r.Close(host)` near the end, or the connection is closed twice.

## Task 2: two pairs

Each thing you acquire gets its own `defer` straight after it. Write them in
the order you acquire things: lock, then connect. Last-in, first-out takes
care of the closing order for you. You can finish with `return r.Run("migrate")`.

## Task 3: the deferred closure

The closure reads and changes the named result `err`. It runs after the
`return` statements have stored their error in `err`:

```go
defer func() {
	if err != nil {
		// replace err with a wrapped version
	}
}()
```

## Almost the answer

```go
func Migrate(r *Recorder, host string) error {
	r.Lock("migrations")
	defer r.Unlock("migrations")
	r.Connect(host)
	defer r.Close(host)
	return r.Run("migrate")
}
```

For `Backup`, inside the closure write
`err = fmt.Errorf("backup %s: %w", host, err)`.
