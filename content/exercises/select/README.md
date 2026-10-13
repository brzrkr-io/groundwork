# select and timeouts

`select` waits on several channels at once and acts on whichever is ready
first. Ops code uses it everywhere: "use whichever replica answers first",
"give up after 5 seconds", "drop the alert rather than block".

## Learn

### select

`select` looks like a `switch`, but every `case` is a channel send or
receive:

```go
select {
case msg := <-primary:
	fmt.Println("primary said", msg)
case msg := <-replica:
	fmt.Println("replica said", msg)
}
```

```text
primary ──┐
          ├──▶ select runs the case that is ready first, once
replica ──┘
```

- `select` waits until at least one case can go ahead, runs that one case,
  and is done. The other cases do not run.
- If several cases are ready at the same moment, Go picks one at random.

### Timeouts with time.After

`time.After(d)` returns a channel that receives a single value once `d` has
passed. Put it in a `select` to stop waiting after a while:

```go
select {
case result := <-results:
	fmt.Println("got", result)
case <-time.After(5 * time.Second):
	fmt.Println("gave up after 5s")
}
```

`case <-ch:` (with nothing on the left) receives a value and ignores it. You
use it when the arrival is the signal and the value doesn't matter.

### Signal channels: chan struct{}

`struct{}` is a type with no fields. It takes no memory and carries no
information, so `chan struct{}` is the usual type for a channel that only
signals "it happened". Usually the sender closes it, which wakes every
receiver at once:

```go
ready := make(chan struct{})
go func() {
	warmUp()
	close(ready) // signal: ready now
}()
<-ready // waits for the close
```

### Not waiting at all: default

A `select` with a `default` case never waits. If no other case is ready
**right now**, `default` runs:

```go
select {
case alerts <- msg:
	// there was room, the alert is queued
default:
	// the buffer is full; don't block, drop it
}
```

### About the tests

The tests use `testing/synctest`, which gives them a fake clock. A
`time.After(5 * time.Second)` fires instantly in real time, but the test still
sees exactly 5 seconds pass on its clock. Your code doesn't need to do
anything special for this.

## Your task

You're writing the parts of a deploy tool that talk to slow, unreliable
services.

### 1. Use whichever replica answers first

`FirstReply(primary, replica)` should wait for an answer on either channel
and return the first one that arrives.

```go
// replica answers "replica: 42 rows" after 1s, primary answers after 2s
FirstReply(primary, replica) // => "replica: 42 rows"   (after 1s)
```

### 2. Wait for a service with a timeout

`WaitReady(ready, timeout)` waits for the `ready` channel to be closed. If
that happens within `timeout`, return `nil` as soon as it does. If `timeout`
passes first, return `ErrNotReady` (already declared in the starter).

```go
// ready is closed after 3s
WaitReady(ready, 5*time.Second) // => nil, after 3s

// ready is never closed
WaitReady(ready, 5*time.Second) // => ErrNotReady, after 5s
```

### 3. Drop alerts instead of blocking

`TryNotify(alerts, msg)` tries to send `msg` on `alerts` without waiting. If
the send can happen right now, do it and return `true`. If it would have to
wait (the buffer is full), return `false` without sending.

```go
alerts := make(chan string, 1)
TryNotify(alerts, "disk 91%") // => true   (now the buffer is full)
TryNotify(alerts, "cpu 99%")  // => false  (dropped)
```

## Learn more

- [A Tour of Go: Select](https://go.dev/tour/concurrency/5)
- [A Tour of Go: Default selection](https://go.dev/tour/concurrency/6)
- [Go by Example: Timeouts](https://gobyexample.com/timeouts)
- [Go by Example: Non-blocking channel operations](https://gobyexample.com/non-blocking-channel-operations)
