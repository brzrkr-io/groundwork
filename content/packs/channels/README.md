# Channels

A *channel* is a pipe that goroutines use to pass values to each other. It is
how a goroutine hands back a result, streams log lines, or queues alerts.

**New here:** `chan`, `make(chan T)`, send `ch <- v`, receive `<-ch`, buffered
channels, `close`, `for range` over a channel.

## Learn

### Making, sending and receiving

```go
ch := make(chan int) // a channel that carries ints

go func() {
	ch <- 200 // send: put a value in
}()

code := <-ch // receive: take a value out (waits until one arrives)
```

The arrow shows which way the value moves: `ch <- v` puts `v` in, `<-ch`
takes a value out. The type says what it carries: `chan int`, `chan string`.

> **Why is it written like this?** Go requires `make` to create a channel;
> a channel declared with `var ch chan int` is `nil` and every send or
> receive on it waits forever. The name `ch` is a convention for short code.

### Unbuffered channels: a handoff

`make(chan T)` is *unbuffered*: it has no storage. A send waits until another
goroutine receives, and the other way round:

```text
sender:    ch <- "a"  ··· waits ···┐
                                   ├─ "a" changes hands, both carry on
receiver:              v := <-ch ──┘
```

So the send and the receive almost always happen in different goroutines.

### Buffered channels: a mailbox

`make(chan T, 3)` is *buffered* with room for 3 values. A send only waits
when the buffer is full; a receive only waits when it is empty:

```text
make(chan string, 3)   [ _ , _ , _ ]
ch <- "a"              [ a , _ , _ ]   didn't wait
ch <- "b"              [ a , b , _ ]   didn't wait
<-ch  → "a"            [ b , _ , _ ]   first in, first out
```

`len(ch)` is how many values are waiting; `cap(ch)` is the buffer size. An
unbuffered channel has `cap(ch) == 0`.

> **Why is it written like this?** The second argument to `make` is the
> capacity. Leave it out (or pass 0) and Go makes an unbuffered channel.

### close and range

When the sender has nothing more to send, it calls `close(ch)`. A `for range`
loop over a channel receives values until the channel is closed **and**
empty, then ends:

```go
for code := range codes {
	fmt.Println("status", code)
}
// we get here once codes is closed and drained
```

| Rule | Why |
|---|---|
| Only the sender closes a channel. | Sending on a closed channel panics. |
| Close when there are no more values. | Otherwise `range` waits forever. |
| You don't have to close every channel. | Only close to say "no more values". |

> **Why is it written like this?** Ranging over a channel gives one variable
> (the value), not two like a slice. Go requires that; there is no index.

## Your task

You're building the plumbing for a log shipper: something that collects
lines, queues alerts and streams lines to the next stage.

### 1. Collect lines from a channel

**Goal:** make `CollectLines(lines)` receive every line until the channel is
closed and return them in the order they arrived.

```go
// another goroutine sends "GET /", "GET /health", then closes lines
CollectLines(lines)
// => []string{"GET /", "GET /health"}
```

**Think it through:**
- Which loop keeps receiving until the channel is closed?
- `collected` already exists. How do you add one line to the end of a slice?

### 2. Queue alerts in a buffered channel

**Goal:** make `QueueAlerts(alerts)` return a **buffered**, **closed** channel
that already holds every alert in order.

Its capacity must be `len(alerts)`. No goroutine is needed: the buffer has
room for every alert, so no send waits.

```go
ch := QueueAlerts([]string{"disk 91%", "cpu 99%"})
cap(ch) // => 2
len(ch) // => 2
// ranging over ch gives "disk 91%", then "cpu 99%", then the loop ends
```

**Think it through:**
- What goes in `make` as the second argument, so there is one slot per alert?
- Three steps in order: make, send each alert, then what before you return?

### 3. Stream lines through an unbuffered channel

**Goal:** make `StreamLines(lines)` return an **unbuffered** channel, with a
goroutine sending each line in order and then closing it.

Return the channel straight away. The goroutine keeps sending while the
caller receives.

```go
ch := StreamLines([]string{"boot", "ready"})
cap(ch) // => 0
for line := range ch {
	fmt.Println(line) // "boot", then "ready"
}
```

If the tests hang and then fail with `panic: test timed out after 60s`, a
send is stuck waiting for a receiver that never comes.

**Think it through:**
- Why would the sends get stuck if they ran in `StreamLines` itself, before
  `return`?
- Which parts go inside the `go func() { ... }()`, and which stay outside?

## Learn more

- [A Tour of Go: Channels](https://go.dev/tour/concurrency/2)
- [A Tour of Go: Buffered channels](https://go.dev/tour/concurrency/3)
- [A Tour of Go: Range and close](https://go.dev/tour/concurrency/4)
- [Go by Example: Channel buffering](https://gobyexample.com/channel-buffering)
