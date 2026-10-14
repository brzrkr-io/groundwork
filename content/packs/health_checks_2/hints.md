## Task 1: Fix the status code

The status and headers go out with the first `Write`, and `Encode` is what calls `Write`. Look at the order of the lines in `WriteReport`.

### Closer

The order must be: set the header, then `w.WriteHeader(code)`, then `json.NewEncoder(w).Encode(report)`.

### Answer

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(code)
json.NewEncoder(w).Encode(report)
```

## Task 2: Run the checks concurrently

Start one goroutine per check and wait for all of them before returning. Don't `append` from goroutines; give each one its own slot in the slice.

### Closer

Make the slice full size up front with `make([]Result, len(checks))`. Use `var wg sync.WaitGroup`, `wg.Go(func() { ... })` inside the loop, and `wg.Wait()` before `return`. Add `"sync"` to the imports.

### Answer

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

Each goroutine writes only its own `results[i]`, so there is no data race and the order matches `checks`.

## Task 3: Gate readiness on startup and draining

The `Mark` methods store `true` in one field each. `State` reads both fields, and draining wins over everything else.

### Closer

Use `g.ready.Store(true)` and `g.draining.Store(true)`. In `State`, a tagless `switch` checks `g.draining.Load()` first, then `!g.ready.Load()`.

### Answer

```go
func (g *Gate) MarkReady() {
	g.ready.Store(true)
}

func (g *Gate) MarkDraining() {
	g.draining.Store(true)
}

func (g *Gate) State() string {
	switch {
	case g.draining.Load():
		return "draining"
	case !g.ready.Load():
		return "starting"
	default:
		return "ready"
	}
}
```
