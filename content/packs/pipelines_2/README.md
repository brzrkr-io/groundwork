# Pipelines: fan-in and cancellation

Two upgrades to the pipelines from *Pipelines*. *Fan-out* runs several copies
of a slow stage at once, and *fan-in* merges their outputs into one channel.
A `context` lets the consumer stop early without leaving goroutines stuck.

**New here:** goroutine leaks, `select` on a send, fan-out, fan-in (`Merge`).

## Learn

### Recap

A stage takes an input, returns an output channel, and works in a goroutine
that runs `defer close(out)`.

### Goroutine leaks

A send `out <- v` waits until someone receives. If the consumer stops
reading early, the stage's goroutine waits **forever**. That is a
*goroutine leak*: memory that is never freed. One per request, and a service
eventually runs out.

### Sending with select

A `select` waits for whichever of its cases can happen first. Put the send
in one case and `ctx.Done()` in the other:

```go
// A ticker that stops when ctx is cancelled.
for n := 1; ; n++ {
	select {
	case ticks <- n: // someone received n
	case <-ctx.Done(): // the consumer cancelled
		return // deferred close(ticks) still runs
	}
}
```

The consumer makes the context with `context.WithCancel` and calls
`cancel()` when it stops reading. That closes `ctx.Done()`, so the second
case fires.

> **Why is it written like this?**
> - **Go requires** each `case` to be one send or one receive, followed by `:`.
>   An empty case body is allowed: after the send, the loop just continues.
> - **Convention:** `ctx` is the first parameter of any function that can be
>   cancelled.
> - **Author's choice:** returning from the goroutine (not `break`, which
>   would only leave the `select`).

### Fan-out and fan-in

```text
                    ┌─▶ check ─┐
hosts ──▶ Generate ─┼─▶ check ─┼─▶ Merge ──▶ results
                    └─▶ check ─┘
           fan-out: 3 copies     fan-in: one channel
           read the same input
```

**Fan-out is free:** several stages can `range` over the **same** channel;
each value goes to exactly one of them.

**Fan-in needs `Merge`.** Its parts:

| Part | Job |
|---|---|
| one output channel | everything is forwarded here |
| one *forwarder* goroutine per input | `range` over its input, send each value to the output |
| a `sync.WaitGroup` | counts the forwarders |
| one *closer* goroutine | `wg.Wait()`, then close the output |

It's the same closer as the `results` channel in *Worker pools*. Merged
values come out in whatever order they arrive.

### Passing a slice to a variadic parameter

`Merge(ctx, inputs ...<-chan string)` takes any number of channels. To pass
a slice of them, add `...` after it, as in *Variadic functions*:

```go
nums := []int{1, 2, 3}
total := sum(nums...)
```

### About the tests

The tests use `testing/synctest`, which has a fake clock. After `cancel()`,
`synctest.Wait()` waits until every goroutine is finished or stuck, then the
test checks that your output is closed and nothing is still trying to send.

## Your task

A fleet checker generates host names, checks them with several workers at
once, and merges the results.

### 1. Fix the leak in Generate

**Goal:** make `Generate`'s send also watch `ctx.Done()`, and stop when it is closed.

`Generate(ctx, lines)` sends every line, then closes its output. After a
cancel, it must stop sending and close its output, even if nobody reads.

```go
ctx, cancel := context.WithCancel(context.Background())
out := Generate(ctx, []string{"web-1", "web-2", "web-3"})
<-out    // "web-1"
cancel() // Generate stops and closes out; "web-2" and "web-3" are never sent
```

**Think it through:**
- Which single line in the loop can block forever?
- What are the two things that line should wait for?

### 2. Merge channels

**Goal:** `Merge(ctx, inputs...)` returns one channel carrying every value from every input.

Values come out in any order. The output is closed once all inputs are
closed and drained. If `ctx` is cancelled, `Merge` stops forwarding and
closes its output, even if nobody is reading. Add `"sync"` to the imports.

```go
Merge(ctx, a, b) // a carries "web-1", "web-2"; b carries "db-1"
// carries "web-1", "web-2" and "db-1" in some order, then is closed
```

**Think it through:**
- Walk the four rows of the fan-in table. Which ones are goroutines?
- Each forwarder's send can block like task 1's. How did you fix that?
- Who closes the output, and when? (Never the forwarders.)

### 3. Fan out

**Goal:** `FanOut(ctx, hosts, workers, check)` starts `workers` copies of `checkStage` and returns `Merge` of their outputs.

Right now it runs one copy. All copies read from the same `hosts` channel.
With 6 hosts, 3 workers and 1-second checks, every result arrives after 2
seconds.

```go
FanOut(ctx, hosts, 3, check) // hosts carries "web-1" ... "web-6"
// carries "web-1 ok" ... "web-6 ok" in some order, then is closed
```

**Think it through:**
- Each `checkStage` call returns one channel. Where will you keep `workers` of them?
- How do you hand a slice to `Merge`'s `inputs ...` parameter?

## Learn more

- [Go blog: Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- [testing/synctest](https://pkg.go.dev/testing/synctest)
- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
