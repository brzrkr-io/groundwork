# Pipelines: fan-in and cancellation

Two upgrades to the pipelines from *Pipelines*. *Fan-out* runs several copies
of a slow stage at the same time, and *fan-in* merges their outputs back into
one channel. And a `context` lets the consumer stop early without leaving
goroutines stuck forever.

## Learn

### Recap

A stage takes an input channel, returns an output channel, and does its
work in a goroutine that closes the output (`defer close(out)`) when the
input runs out.

### Goroutine leaks

What if the consumer stops reading early, say after finding the first error?

```go
for _, line := range lines {
	out <- line // nobody is receiving any more: blocks forever
}
```

The stage's goroutine is stuck on that send for the rest of the program's
life. That is a *goroutine leak*. One leak is small, but a service that
leaks one per request eventually runs out of memory. The fix: every send also
watches a context, so cancelling stops the stage:

```go
for _, line := range lines {
	select {
	case out <- line:
	case <-ctx.Done():
		return // the consumer gave up; deferred close(out) still runs
	}
}
```

The consumer creates the context and calls `cancel()` when it stops reading.

### Fan-out and fan-in

```text
                    ┌─▶ check ─┐
hosts ──▶ Generate ─┼─▶ check ─┼─▶ Merge ──▶ results
                    └─▶ check ─┘
           fan-out: 3 copies     fan-in: one channel
           read the same input
```

Fan-out is free: several stages can `range` over the **same** input
channel, and each value goes to exactly one of them. Fan-in needs a `Merge`
function. It starts one forwarding goroutine per input. Each forwarder copies
values to a shared output. A WaitGroup tracks the forwarders, and one more
goroutine closes the output when they have all finished, just like the
results channel of a worker pool:

```go
out := make(chan string)
var wg sync.WaitGroup
for _, in := range inputs {
	wg.Go(func() {
		for v := range in {
			// forward v to out (and watch ctx.Done(), as above)
		}
	})
}
go func() {
	wg.Wait()
	close(out)
}()
return out
```

Merged values come out in whatever order they arrive.

### About the tests

The tests use `testing/synctest`. `synctest.Wait()` waits until every
goroutine in the test is either finished or stuck. The tests use it after
`cancel()` to check that your goroutines really stopped.

## Your task

A fleet checker generates host names, checks them with several workers at
once, and merges the results.

### 1. Fix the leak in Generate

`Generate(ctx, lines)` sends every line on its output, then closes it. If
the consumer cancels `ctx` and stops reading, the goroutine is stuck on its
send forever. Make every send also watch `ctx.Done()`. After a cancel,
`Generate` must stop sending and close its output.

```go
ctx, cancel := context.WithCancel(context.Background())
out := Generate(ctx, []string{"web-1", "web-2", "web-3"})
<-out    // "web-1"
cancel() // Generate stops and closes out; "web-2" and "web-3" are never sent
```

### 2. Merge channels

`Merge(ctx, inputs...)` returns one channel that carries every value from
every input, in any order. It is closed once all inputs are closed and
drained. If `ctx` is cancelled, `Merge` stops forwarding and closes its
output, even if nobody is reading.

```go
Merge(ctx, a, b) // a carries "web-1", "web-2"; b carries "db-1"
// carries "web-1", "web-2" and "db-1" in some order, then is closed
```

### 3. Fan out

`FanOut(ctx, hosts, workers, check)` runs just one copy of the provided
`checkStage` so far. Make it start `workers` copies, all reading from the same
`hosts` channel, and return `Merge` of their outputs. With 6 hosts, 3 workers
and 1-second checks, all results arrive after 2 seconds.

```go
FanOut(ctx, hosts, 3, check) // hosts carries "web-1" ... "web-6"
// carries "web-1 ok" ... "web-6 ok" in some order, then is closed
```

## Learn more

- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [testing/synctest](https://pkg.go.dev/testing/synctest)
- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
