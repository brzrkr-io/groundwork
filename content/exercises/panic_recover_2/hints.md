## Task 1: Make Guard stop the panic

Right now `recover()` runs before the handler, when nothing has gone wrong
yet. Only deferred functions run during a panic, so that's where it belongs.

### Closer

Wrap the existing `if r := recover(); r != nil { ... }` in
`defer func() { ... }()`, and place it before `err = handler()`.

### Answer

```go
var err error
defer func() {
	if r := recover(); r != nil {
		err = fmt.Errorf("command %q panicked: %v", command, r)
	}
}()
err = handler()
return err
```

The panic is now stopped, but `Guard` still returns `nil` after one; task 2
fixes that.

## Task 2: Return the panic as an error

After a panic the `return` line never runs, so setting a local `err` is lost.
The deferred function must set the function's result itself.

### Closer

Change the signature to
`func Guard(command string, handler func() error) (err error)` and delete
`var err error`. The last two lines can become `return handler()`.

### Answer

```go
func Guard(command string, handler func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("command %q panicked: %v", command, r)
		}
	}()
	return handler()
}
```

A named result is what the caller receives, so the deferred function's
assignment survives the panic.

## Task 3: Fix the real bug instead of recovering

`var counts map[string]int` declares a map but never creates one. Writing to
a `nil` map always panics, and the `recover` just hides that.

### Closer

Delete the whole `defer func() { recover() }()` block, and replace the `var`
line with `counts := make(map[string]int)`.

### Answer

```go
counts := make(map[string]int)
for _, a := range alerts {
	counts[a.Severity]++
}
return counts
```

`make` creates an empty map, so no alerts gives `map[string]int{}`, not `nil`.
