# Channels: deadlocks and directions

Channel bugs usually show up as a program that freezes. Here you'll see what
that looks like, fix it, and use two tools that make channel code safer.

**New here:** deadlocks, direction types `<-chan T` and `chan<- T`, the
comma-ok receive `v, ok := <-ch`.

## Learn

### Recap

`make(chan T)` is unbuffered: a send waits for a receiver. `make(chan T, n)`
is buffered: a send only waits when `n` values are already waiting. The
sender calls `close(ch)` when done; `for v := range ch` stops after the
channel is closed and empty.

### Deadlocks

If every goroutine is waiting on a channel and nobody can move, that is a
*deadlock*. Go crashes with:

```text
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/home/you/tool/main.go:9 +0x37
```

- `all goroutines are asleep`: nobody can make progress.
- `[chan send]`: goroutine 1 is stuck **sending**. You may also see
  `[chan receive]` or `[sync.WaitGroup.Wait]`.
- `main.go:9`: the line it is stuck on. The fix is often elsewhere, such as
  where the channel is made.

In a test you usually see a hang or a timeout instead. Ask: *who is supposed
to be on the other end of this channel?*

### Direction types

A function can promise to only receive from, or only send to, a channel:

| Type | Meaning |
|---|---|
| `chan string` | send and receive |
| `<-chan string` | receive only (arrow points out of `chan`) |
| `chan<- string` | send only (arrow points into `chan`) |

```go
func Double(in <-chan int, out chan<- int) {
	for n := range in {
		out <- n * 2
	}
	close(out)
}
```

> **Why is it written like this?** Go requires the arrow's position: it is
> part of the type. A plain `chan int` converts automatically when you pass
> it. Sending on `in` would not compile: `invalid operation: cannot send to
> receive-only channel in`. Using direction types at all is a convention
> that documents who sends and who receives.

### The comma-ok receive

`v, ok := <-ch` tells you whether the value is real. `ok` is `false` when the
channel is closed and empty; `v` is then the zero value (`""` for strings):

```go
port, ok := <-ports
if !ok {
	fmt.Println("no more ports")
}
```

> **Why is it written like this?** Go requires exactly this form for the
> second result. The name `ok` is a convention.

## Your task

You're cleaning up a teammate's log-shipping code.

### 1. Fix the deadlock

**Goal:** make `CheckHosts(hosts, check)` stop freezing, so it returns one
answer per host, in any order.

Each goroutine sends its answer and waits for a receiver. But the main
goroutine only receives after `wg.Wait()` returns, so nobody can move. Make
every send complete without waiting for a receiver. The test sorts the
answers.

```go
CheckHosts([]string{"web-1", "web-2"}, check)
// => []string{"web-1 ok", "web-2 ok"}   (in any order)
```

**Think it through:**
- The deadlock shows up at the send, but which line *creates* the channel?
- How many answers will be sent in total? How much room does that need?

### 2. Prefix lines with direction types

**Goal:** make `PrefixLines(host, in, out)` send each line from `in` to `out`
as `host: line`, until `in` is closed.

The text sent is: the `host` parameter, then the fixed text `": "` (colon and
a space), then the line. The starter already closes `out` at the end; keep it.

```go
// in carries "boot", "ready", then is closed
PrefixLines("web-1", in, out)
// out carries "web-1: boot", "web-1: ready", then is closed
```

**Think it through:**
- Which loop receives from `in` until it is closed?
- Three pieces are joined with `+`. Which ones come from variables, and which
  is text in quotes? Don't forget the space.

### 3. Take a few jobs

**Goal:** make `Take(jobs, n)` receive up to `n` jobs and return them in
order, plus `closed` saying whether the channel ran out first.

- If the channel is closed before you have `n` jobs: stop and return what you
  have, with `closed` = `true`.
- If you get all `n` jobs: `closed` is `false`, even if the channel happens
  to be closed right after them.

```go
// jobs holds "a", "b", "c" and is open
Take(jobs, 2) // => []string{"a", "b"}, false

// jobs holds "a", "b", "c" and is closed
Take(jobs, 5) // => []string{"a", "b", "c"}, true
Take(jobs, 3) // => []string{"a", "b", "c"}, false
```

**Think it through:**
- Why not `for range jobs`? It would keep going past `n`, and on an open
  channel it never stops. Loop `n` times instead.
- Each time round, use a comma-ok receive. What do you return when `ok` is
  `false`?

## Learn more

- [Go spec: Channel types](https://go.dev/ref/spec#Channel_types)
- [Go spec: Receive operator](https://go.dev/ref/spec#Receive_operator)
- [Go by Example: Channel directions](https://gobyexample.com/channel-directions)
