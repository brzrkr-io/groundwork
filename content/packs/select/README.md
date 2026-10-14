# select and timeouts

`select` waits on several channels at once and acts on whichever is ready
first. Ops code uses it for "use whichever replica answers first", "give up
after 5 seconds" and "drop the alert rather than block".

**New here:** `select`, `time.After`, `chan struct{}`, `default` in a select.

## Learn

### select

`select` looks like a `switch`, but every `case` is a channel send or
receive:

```go
select {
case n := <-eu:
	fmt.Println("eu region:", n)
case n := <-us:
	fmt.Println("us region:", n)
}
```

```text
eu ──┐
     ├──▶ select runs the case that is ready first, once
us ──┘
```

- `select` waits until at least one case can go ahead, runs that one case,
  and is done. The other cases don't run.
- If several are ready at the same moment, Go picks one at random.

> **Why is it written like this?** Go requires every `case` in a `select` to
> be a send or a receive. `n := <-eu` receives and names the value; the
> variable only exists inside that case.

### Timeouts with time.After

`time.After(d)` returns a channel that receives one value once `d` has
passed. Put it in a `select` to stop waiting:

```go
select {
case size := <-sizes:
	fmt.Println("size", size)
case <-time.After(2 * time.Second):
	fmt.Println("gave up after 2s")
}
```

`case <-ch:` (nothing on the left) receives and ignores the value. Use it
when the arrival itself is the signal.

### Signal channels: chan struct{}

`struct{}` is a type with no fields. It carries no information, so
`chan struct{}` is the usual type for "it happened". The sender usually
closes it, which wakes every receiver at once:

```go
done := make(chan struct{})
go func() {
	migrate()
	close(done) // signal: finished
}()
<-done // waits for the close
```

> **Why is it written like this?** Using `struct{}` for signals is a
> convention; `chan bool` would also work but suggests the value matters.
> A receive on a closed channel never waits; Go guarantees that.

### Not waiting at all: default

A `select` with a `default` case never waits. If no other case is ready
**right now**, `default` runs:

```go
select {
case n := <-jobs:
	fmt.Println("got job", n)
default:
	fmt.Println("no job waiting; doing something else")
}
```

### About the tests

The tests use `testing/synctest`, which gives them a fake clock. A 5-second
`time.After` fires instantly in real time, but the test still sees exactly 5
seconds pass. Your code needs nothing special for this.

## Your task

You're writing the parts of a deploy tool that talk to slow, unreliable
services.

### 1. Use whichever replica answers first

**Goal:** make `FirstReply(primary, replica)` wait on both channels and
return the first answer that arrives.

The starter only listens to `primary`, so it waits even when `replica` is
faster.

```go
// replica answers "replica: 42 rows" after 1s, primary after 2s
FirstReply(primary, replica) // => "replica: 42 rows"   (after 1s)
```

**Think it through:**
- How many cases does the `select` need, and which channel is in each?
- What should each case do with the value it received?

### 2. Wait for a service with a timeout

**Goal:** make `WaitReady(ready, timeout)` return `nil` as soon as `ready` is
closed, or `ErrNotReady` if `timeout` passes first.

`ErrNotReady` is already declared in the starter; return it, don't create a
new error. Use the `timeout` parameter, not a fixed number of seconds.

```go
// ready is closed after 3s
WaitReady(ready, 5*time.Second) // => nil, after 3s

// ready is never closed
WaitReady(ready, 5*time.Second) // => ErrNotReady, after 5s
```

**Think it through:**
- Which two things are you waiting for? Each becomes one case.
- Does the `ready` case need the received value, or only the fact it arrived?

### 3. Drop alerts instead of blocking

**Goal:** make `TryNotify(alerts, msg)` send `msg` on `alerts` only if that
can happen right now, and report whether it did.

If the send can happen, do it and return `true`. If it would have to wait
(the buffer is full), return `false` without sending.

```go
alerts := make(chan string, 1)
TryNotify(alerts, "disk 91%") // => true   (now the buffer is full)
TryNotify(alerts, "cpu 99%")  // => false  (dropped)
```

**Think it through:**
- A `case` can be a *send*, not only a receive. Which way does the arrow go?
- Which kind of case runs when nothing else is ready right now?

## Learn more

- [A Tour of Go: Select](https://go.dev/tour/concurrency/5)
- [A Tour of Go: Default selection](https://go.dev/tour/concurrency/6)
- [Go by Example: Timeouts](https://gobyexample.com/timeouts)
- [Go by Example: Non-blocking channel operations](https://gobyexample.com/non-blocking-channel-operations)
