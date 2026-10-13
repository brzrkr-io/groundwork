# Signals: graceful stop for a rolling drain

When you press Ctrl-C, or Kubernetes stops a pod, your program receives a
*signal*. By default a Go program dies on the spot. Here you make a rolling
maintenance job stop *gracefully*: finish the current host, then stop.

**New here:** signals, `signal.Notify`, channels and contexts (just enough).

## Learn

### What signals are

A signal is a tiny message the operating system sends a running program:

| Signal | Who sends it | Can you catch it? |
|---|---|---|
| `SIGINT` | Ctrl-C in your terminal | yes |
| `SIGTERM` | `kill <pid>`, `docker stop`, Kubernetes deleting a pod | yes |
| `SIGKILL` | `kill -9`, or Kubernetes when time runs out | **no** |

Kubernetes sends `SIGTERM`, waits a *grace period* (30 s by default), then
sends `SIGKILL`. That's your window to stop cleanly.

Why bother? A job killed halfway through draining a host may leave it out of
the load balancer but unpatched. Stopping *between* hosts leaves everything in
a known state.

### Exit codes for signals

By Unix convention, a process stopped by signal number *n* exits with
`128 + n`. `SIGINT` is 2, so `130`. Scripts use this to tell "interrupted"
apart from "failed" (`1`).

### Channels, just enough

`os/signal` delivers signals on a *channel*: a pipe that values arrive on.

```go
sigs := make(chan os.Signal, 1)                    // room for 1 waiting value
signal.Notify(sigs, os.Interrupt, syscall.SIGTERM) // deliver these to sigs

for sig := range sigs { // waits for each value; ends when the channel closes
	fmt.Println("got", sig)
}
```

`os.Interrupt` is `SIGINT`. With `%v`, `os.Interrupt` prints as `interrupt`
and `syscall.SIGTERM` as `terminated`.

> **Why is it written like this?** `<-chan os.Signal` in a function line means
> "a channel you can only receive from". Go enforces it. `sig` in the `for`
> line is a name you choose.

### Contexts, just enough

A `context.Context` carries a "please stop" message into your code (the
*Concurrency* module covers it properly). One rule for now:
**`ctx.Err() != nil` means you've been asked to stop.**

```go
ctx, cancel := context.WithCancel(context.Background())
fmt.Println(ctx.Err()) // <nil>
cancel()
fmt.Println(ctx.Err()) // context canceled  (this is context.Canceled)
```

### The full program

This is the `main` you're building towards. `go WatchSignals(...)` runs it in
the background while `Drain` works; that's all you need to know about `go`.

```go
func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go signals.WatchSignals(sigs, cancel, func() { os.Exit(130) }, os.Stderr)

	_, err := signals.Drain(ctx, os.Args[1:], drainHost, os.Stderr)
	os.Exit(signals.ExitCode(err))
}
```

The tests never send real signals. They cancel contexts and fill channels
directly, so they're instant and repeatable.

## Your task

Your maintenance job drains hosts one at a time, and must never stop halfway
through a host.

### 1. Stop between hosts

**Goal:** in `Drain`, at the **start** of each loop step, stop if `ctx` has
been cancelled.

When stopping, write this line to `log`, then return the number drained so
far and `ctx.Err()`:

```text
stopping: <N> hosts not drained
```

`N` is the number of hosts not drained yet (a number, then fixed text, then a
newline). The host in progress always finishes.

```go
// ctx gets cancelled while web-2 is being drained
n, err := Drain(ctx, []string{"web-1", "web-2", "web-3", "web-4"}, drain, log)
// n == 2, errors.Is(err, context.Canceled) == true
// log:
// drained web-1
// drained web-2
// stopping: 2 hosts not drained
```

Already cancelled before the first of 3 hosts: nothing drained, log is just
`stopping: 3 hosts not drained`.

**Think it through:**
- Where's the `TODO(1)`? It's before `drain(host)` on purpose.
- You have `len(hosts)` and the `drained` counter. How do you get "not drained"?

### 2. Pick the exit code

**Goal:** make `ExitCode(err)` return `0`, `130` or `1`.

| `err` | Exit code |
|---|---|
| `nil` | `0` |
| `context.Canceled`, or an error wrapping it | `130` |
| anything else | `1` |

```go
ExitCode(nil)                                       // => 0
ExitCode(fmt.Errorf("drain: %w", context.Canceled)) // => 130
ExitCode(errors.New("disk full"))                   // => 1
```

Add `"errors"` to the imports.

**Think it through:**
- The error may be wrapped, so `==` won't find it. What did *Errors* use?
- Which case must be checked first?

### 3. Watch for signals

**Goal:** write `WatchSignals`: the first signal cancels, the second forces.

Loop over `sigs` with `range`:

- **First** signal: write
  `received <sig>: finishing current host (send again to force quit)`
  to `log`, then call `cancel()`.
- **Second** signal: write `received <sig> again: forcing exit` to `log`,
  call `force()`, and return. (A third signal is never handled.)
- Channel closes: return.

`<sig>` is the signal printed with `%v`. Each line ends with a newline.

```text
received interrupt: finishing current host (send again to force quit)
received interrupt again: forcing exit
```

For `syscall.SIGTERM` the word is `terminated`. With no signals, nothing is
logged and neither function is called.

**Think it through:**
- How will the loop remember it already saw one signal? A variable outside it.
- `cancel` and `force` are parameters that hold functions. How do you call one?

## Learn more

- [os/signal package](https://pkg.go.dev/os/signal)
- [Go by Example: Signals](https://gobyexample.com/signals)
- [Kubernetes: Termination of Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
- [context package](https://pkg.go.dev/context)
