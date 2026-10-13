## Task 1: swap two lines

The body is written by `Encode`. The status has to be sent before it, so
`w.WriteHeader(code)` must come before `json.NewEncoder(w).Encode(report)`.

## Task 2: one slot per check

Don't `append` from goroutines. Make the slice full size up front with
`make([]Result, len(checks))`, then let goroutine `i` fill `results[i]`.
Call `wg.Wait()` before returning.

## Task 2: almost the answer

```go
results := make([]Result, len(checks))
var wg sync.WaitGroup
for i, c := range checks {
	wg.Go(func() {
		results[i] = runCheck(ctx, c, timeout)
	})
}
wg.Wait()
return results
```

## Task 3: the gate

`MarkReady` is `g.ready.Store(true)`. In `State`, check `g.draining.Load()`
first, then `g.ready.Load()`:

```go
switch {
case g.draining.Load():
	return "draining"
case !g.ready.Load():
	return "starting"
default:
	return "ready"
}
```
