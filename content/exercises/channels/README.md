# Channels

A *channel* is a pipe that goroutines use to pass values to each other. It is
how a goroutine hands back a result, streams log lines, or queues alerts for
another part of your program.

## Learn

### Making, sending and receiving

```go
ch := make(chan string) // a channel that carries strings

go func() {
	ch <- "deploy finished" // send: put a value into the channel
}()

msg := <-ch // receive: take a value out (waits until one arrives)
fmt.Println(msg)
```

The arrow shows which way the value moves: `ch <- v` puts `v` in, `<-ch`
takes a value out. A channel's type says what it carries: `chan string`,
`chan int`, `chan Alert`.

### Unbuffered channels: a handoff

A channel made with `make(chan string)` is *unbuffered*. It has no storage, so
a send waits until another goroutine receives, and a receive waits until
another goroutine sends. The two meet, the value changes hands, and both carry
on:

```text
sender:    ch <- "a"  ··· waits ···┐
                                   ├─ "a" changes hands, both carry on
receiver:              v := <-ch ──┘
```

Because the sender waits for a receiver, a send on an unbuffered channel
almost always happens in a different goroutine from the receive.

### Buffered channels: a mailbox

`make(chan string, 3)` makes a *buffered* channel with room for 3 values. A send
only waits when the buffer is full, and a receive only waits when it is empty:

```text
make(chan string, 3)      [ _ , _ , _ ]
ch <- "a"                 [ a , _ , _ ]   didn't wait
ch <- "b"                 [ a , b , _ ]   didn't wait
<-ch  → "a"               [ b , _ , _ ]   first in, first out
```

`len(ch)` is how many values are waiting in the buffer, and `cap(ch)` is its
size. An unbuffered channel has `cap(ch) == 0`.

### close and range

When the sender has nothing more to send, it calls `close(ch)`. A `for range`
loop over a channel receives values until the channel is closed **and** empty,
then the loop ends:

```go
go func() {
	for _, line := range []string{"GET /", "GET /health"} {
		ch <- line
	}
	close(ch) // tell the receiver there is nothing more
}()

for line := range ch {
	fmt.Println(line)
}
// we get here once ch is closed and drained
```

Rules of thumb:

| Rule | Why |
|---|---|
| Only the sender closes a channel. | Sending on a closed channel panics. |
| Close when there are no more values. | Otherwise `range` waits forever. |
| You don't have to close every channel. | Only close to say "no more values". |

## Your task

You're building the plumbing for a log shipper: something that collects lines,
queues alerts and streams lines to the next stage.

### 1. Collect lines from a channel

`CollectLines(lines)` should receive every line from the `lines` channel until
it is closed, and return them in the order they arrived.

```go
// another goroutine sends "GET /", "GET /health", then closes lines
CollectLines(lines)
// => []string{"GET /", "GET /health"}
```

### 2. Queue alerts in a buffered channel

`QueueAlerts(alerts)` should return a **buffered** channel whose capacity is
`len(alerts)`, already holding every alert in order, and already **closed**. No
goroutine is needed, because the buffer has room for every alert.

```go
ch := QueueAlerts([]string{"disk 91%", "cpu 99%"})
cap(ch) // => 2
len(ch) // => 2
// ranging over ch gives "disk 91%", then "cpu 99%", then the loop ends
```

### 3. Stream lines through an unbuffered channel

`StreamLines(lines)` should return an **unbuffered** channel and start a
goroutine that sends each line in order, then closes the channel. Return the
channel straight away. The goroutine keeps sending while the caller receives.

```go
ch := StreamLines([]string{"boot", "ready"})
cap(ch) // => 0
for line := range ch {
	fmt.Println(line) // "boot", then "ready"
}
```

## Learn more

- [A Tour of Go: Channels](https://go.dev/tour/concurrency/2)
- [A Tour of Go: Buffered channels](https://go.dev/tour/concurrency/3)
- [A Tour of Go: Range and close](https://go.dev/tour/concurrency/4)
- [Go by Example: Channel buffering](https://gobyexample.com/channel-buffering)
