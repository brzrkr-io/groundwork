## Task 1: Warm caches in the background

Right now each `warm(host)` call runs and finishes before the loop moves on. One keyword in front of a call makes it run in its own goroutine.

### Closer

Put `go` in front of the call. Nothing else changes, and you don't wait for anything:
`go someCall(arg)`.

### Answer

```go
for _, host := range hosts {
	go warm(host)
}
```

`go` starts the call in a new goroutine and carries on at once, so `WarmCaches` returns straight away.

## Task 2: Restart services and wait

You need to start every restart at the same time and then wait for all of them. A `sync.WaitGroup` does both; remember to add `"sync"` to the imports.

### Closer

Create `var wg sync.WaitGroup` before the loop, start each restart with `wg.Go(func() { ... })` inside the loop, and call `wg.Wait()` once **after** the loop.

### Answer

```go
import "sync"

var wg sync.WaitGroup
for _, service := range services {
	wg.Go(func() {
		restart(service)
	})
}
wg.Wait()
```

All restarts start before `wg.Wait()`, so they overlap and the whole thing takes about as long as one restart.
