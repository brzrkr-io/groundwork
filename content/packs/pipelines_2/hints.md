## Task 1: Fix the leak in Generate

`out <- line` waits until someone receives. If the consumer has cancelled and
stopped reading, it waits forever. Make the send wait for either the receiver
or `ctx.Done()`, whichever comes first.

### Closer

Replace the plain send with a `select` that has two cases: `case out <- line:`
and `case <-ctx.Done():`. In the second case, `return`.

### Answer

```go
for _, line := range lines {
	select {
	case out <- line:
	case <-ctx.Done():
		return
	}
}
```

`return` ends the goroutine, and the deferred `close(out)` still runs.

## Task 2: Merge channels

Follow the fan-in sketch in the README: one forwarding goroutine per input, a
WaitGroup to track them, and one more goroutine that closes the output when
they are all done. Each forwarder's send should watch `ctx.Done()`, as in task 1.

### Closer

Make `out` and a `sync.WaitGroup` (add `"sync"` to the imports). For each
input, `wg.Go` a forwarder that does `for v := range in` and sends `v` with
the same `select` as task 1. Then `go func() { wg.Wait(); close(out) }()`
and `return out`.

### Answer

```go
out := make(chan string)
var wg sync.WaitGroup
for _, in := range inputs {
	wg.Go(func() {
		for v := range in {
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	})
}
go func() {
	wg.Wait()
	close(out)
}()
return out
```

Only the closer goroutine closes `out`, and only after every forwarder has stopped.

## Task 3: Fan out

Fan-out needs nothing special: several `checkStage`s can read from the same
`hosts` channel. Start `workers` of them, collect their output channels, and
merge those with `Merge`.

### Closer

Make a slice `outputs := make([]<-chan string, workers)` and fill each slot
with `checkStage(ctx, hosts, check)`. Pass the slice to `Merge` with
`outputs...`.

### Answer

```go
outputs := make([]<-chan string, workers)
for i := range outputs {
	outputs[i] = checkStage(ctx, hosts, check)
}
return Merge(ctx, outputs...)
```

`outputs...` passes the slice to `Merge`'s variadic parameter, as in *Variadic functions*.
