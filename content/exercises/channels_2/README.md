# Channels: deadlocks and directions

Channel bugs usually show up as a program that freezes. This exercise shows
you what that looks like, how to fix it, and two tools that make channel code
safer: direction types and the "comma ok" receive.

## Learn

### Recap

`make(chan T)` is unbuffered: a send waits for a receiver. `make(chan T, n)`
is buffered: a send only waits when `n` values are already waiting. The sender
calls `close(ch)` when it is done, and `for v := range ch` stops after the
channel is closed and empty.

### Deadlocks

If every goroutine is waiting on a channel and nobody can ever move, the
program is stuck. That is a *deadlock*. Go spots it and crashes with:

```text
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [chan send]:
main.main()
	/home/you/tool/main.go:9 +0x37
```

- `all goroutines are asleep`: nobody can make progress.
- `goroutine 1 [chan send]`: goroutine 1 (main) is stuck **sending**. You
  may also see `[chan receive]` or `[sync.WaitGroup.Wait]`.
- `main.go:9`: the exact line it is stuck on. Start there.

Inside a test you usually see a hang instead, or a timeout from the test
itself. The fix always starts with the question: *who is supposed to be on
the other end of this channel?*

### Direction types

A function can promise to only receive from, or only send to, a channel:

| Type | Meaning |
|---|---|
| `chan string` | send and receive |
| `<-chan string` | receive only (the arrow points out of `chan`) |
| `chan<- string` | send only (the arrow points into `chan`) |

```go
func Forward(in <-chan string, out chan<- string) {
	for v := range in {
		out <- v
	}
	close(out)
}
```

A plain `chan string` converts automatically when you pass it to these
parameters. If `Forward` tried to send on `in`, the compiler would stop you:
`invalid operation: cannot send to receive-only channel in`.

### The comma-ok receive

`v, ok := <-ch` tells you whether the value is real. `ok` is `false` when
the channel is closed and empty, and `v` is then the zero value (`""` for
strings):

```go
job, ok := <-jobs
if !ok {
	fmt.Println("no more jobs")
}
```

## Your task

You're cleaning up a teammate's log-shipping code.

### 1. Fix the deadlock

`CheckHosts(hosts, check)` runs `check(host)` for every host in its own
goroutine and returns all the answers, in any order. Right now it freezes
forever: each goroutine waits for someone to receive its answer, while the
main goroutine waits in `wg.Wait()` for the goroutines to finish.

Fix it so every send can complete without a receiver. The test sorts the
answers, so order doesn't matter.

```go
CheckHosts([]string{"web-1", "web-2"}, check)
// => []string{"web-1 ok", "web-2 ok"}   (in any order)
```

### 2. Prefix lines with direction types

`PrefixLines(host, in, out)` should receive every line from `in` until it is
closed, send `host + ": " + line` to `out` for each one, and then close
`out`. The signature already uses direction types; the starter already closes
`out`.

```go
// in carries "boot", "ready", then is closed
PrefixLines("web-1", in, out)
// out carries "web-1: boot", "web-1: ready", then is closed
```

### 3. Take a few jobs

`Take(jobs, n)` should receive up to `n` jobs and return them in order. Use a
comma-ok receive: if the channel turns out to be closed before you have `n`
jobs, stop and return what you have, with `closed` set to `true`. If you get
all `n` jobs, `closed` is `false`, even if the channel happens to be closed
right after them.

```go
// jobs holds "a", "b", "c" and is open
Take(jobs, 2) // => []string{"a", "b"}, false

// jobs holds "a", "b", "c" and is closed
Take(jobs, 5) // => []string{"a", "b", "c"}, true
Take(jobs, 3) // => []string{"a", "b", "c"}, false
```

## Learn more

- [Go spec: Channel types](https://go.dev/ref/spec#Channel_types)
- [Go spec: Receive operator](https://go.dev/ref/spec#Receive_operator)
- [Go by Example: Channel directions](https://gobyexample.com/channel-directions)
