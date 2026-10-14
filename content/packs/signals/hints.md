## Task 1: Stop between hosts

The check belongs at the very top of the loop body, before `drain(host)` is
called. That way the host in progress always finishes.

### Closer

```go
if ctx.Err() != nil {
	// log, then return
}
```

The hosts not yet drained are `len(hosts) - drained`.

### Answer

```go
if ctx.Err() != nil {
	fmt.Fprintf(log, "stopping: %d hosts not drained\n", len(hosts)-drained)
	return drained, ctx.Err()
}
```

## Task 2: Pick the exit code

Check the cases in order: no error, a cancelled context, anything else. The
error may be wrapped, so plain `==` won't find `context.Canceled`.

### Closer

`errors.Is(err, context.Canceled)` looks through wrapping. A `switch` with no
value after it (`switch { case err == nil: ... }`) reads nicely here. Add
`"errors"` to the imports.

### Answer

```go
switch {
case err == nil:
	return 0
case errors.Is(err, context.Canceled):
	return 130
default:
	return 1
}
```

## Task 3: Watch for signals

You need to remember whether you've already seen a signal. Keep a variable
outside the loop and check it each time a signal arrives.

### Closer

Use `first := true` before `for sig := range sigs`. The first time round, log,
call `cancel()`, set `first = false` and `continue`. Any later time, log, call
`force()` and `return`.

### Answer

```go
first := true
for sig := range sigs {
	if first {
		fmt.Fprintf(log, "received %v: finishing current host (send again to force quit)\n", sig)
		cancel()
		first = false
		continue
	}
	fmt.Fprintf(log, "received %v again: forcing exit\n", sig)
	force()
	return
}
```
