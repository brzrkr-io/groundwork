## Task 1: check before each host

Put an `if ctx.Err() != nil { ... }` as the very first thing inside the loop,
before `drain(host)` is called. The hosts not yet drained are
`len(hosts) - drained`.

## Task 2: errors.Is

Check `err == nil` first, then `errors.Is(err, context.Canceled)`. A `switch`
with no value after it (`switch { case ...: }`) reads nicely here. Remember to
add `"errors"` to the imports.

## Task 3: remember the first signal

Use a variable such as `first := true` outside the loop. Inside
`for sig := range sigs`, the first time round you log, call `cancel()` and set
`first = false`. Any later time you log, call `force()` and `return`.

## Almost the answer

```go
if ctx.Err() != nil {
	fmt.Fprintf(log, "stopping: %d hosts not drained\n", len(hosts)-drained)
	return drained, ctx.Err()
}
```

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
