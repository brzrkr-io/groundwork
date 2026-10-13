# Signals: graceful stop for a rolling drain

When you press Ctrl-C, or Kubernetes stops a pod, your program receives a
*signal*. By default a Go program dies on the spot, even in the middle of
work. Here you make a rolling maintenance job stop *gracefully* instead.

## Learn

### What signals are

A signal is a small message the operating system delivers to a running
process. Three of them matter most for ops tools:

| Signal | Who sends it | Can you catch it? |
|---|---|---|
| `SIGINT` | Ctrl-C in your terminal | yes |
| `SIGTERM` | `kill <pid>`, `docker stop`, Kubernetes deleting a pod | yes |
| `SIGKILL` | `kill -9`, or Docker/Kubernetes when time runs out | **no** |

When Kubernetes terminates a pod it sends `SIGTERM`, waits for the *grace
period* (30 seconds by default), and then sends `SIGKILL`. So you have a short
window to finish up cleanly.

Why bother? If a maintenance job dies halfway through draining a host, that
host may be left half-drained: out of the load balancer but not yet patched.
A file may be half-written. A graceful stop finishes the current step and then
stops between steps, where everything is in a known state.

### Exit codes for signals

By Unix convention, a process that stops because of signal number *n* exits
with code `128 + n`. `SIGINT` is signal 2, so the code is `130`. `SIGTERM` is
signal 15, so the code is `143`. Scripts and CI systems use these codes to
tell "interrupted" apart from "failed" (`1`).

### Channels, just enough

The `os/signal` package delivers signals on a *channel*. Think of a channel as
a pipe that values arrive on. You make one with `make`, and the `1` gives it
room to hold one value that nobody has read yet:

```go
sigs := make(chan os.Signal, 1)
signal.Notify(sigs, os.Interrupt, syscall.SIGTERM) // deliver these to sigs

for sig := range sigs { // waits for each value, ends when the channel is closed
	fmt.Println("got", sig)
}
```

`os.Interrupt` is `SIGINT`. When printed with `%v`, `os.Interrupt` shows as
`interrupt` and `syscall.SIGTERM` shows as `terminated`. You need to import
`os/signal` and `syscall`. The type `<-chan os.Signal` means "a channel you
can only receive from".

### Contexts, just enough

A `context.Context` carries a "please stop" message into your code.
(The *Concurrency* module covers contexts properly.) For now you need one
rule: **`ctx.Err() != nil` means you have been asked to stop.** After a
cancel, `ctx.Err()` returns `context.Canceled`.

```go
ctx, cancel := context.WithCancel(context.Background())
fmt.Println(ctx.Err()) // <nil>
cancel()
fmt.Println(ctx.Err()) // context canceled
```

The simplest way to connect signals to a context is `signal.NotifyContext`.
It cancels the context on the first signal:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
```

When you want to react to *each* signal yourself (for example, "first Ctrl-C
stops gracefully, second Ctrl-C forces quit"), use `signal.Notify` with your
own channel and call `cancel` yourself. That is what this exercise does.

### The full program

This is the `main` you are building towards. `go WatchSignals(...)` runs
`WatchSignals` in the background while `Drain` does the work. That one line is
all you need to know about `go` for now.

```go
func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go signals.WatchSignals(sigs, cancel, func() { os.Exit(130) }, os.Stderr)

	_, err := signals.Drain(ctx, os.Args[1:], drainHost, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "drain: %v\n", err)
	}
	os.Exit(signals.ExitCode(err))
}
```

The tests never send real signals. They cancel contexts and fill channels
directly, so everything runs instantly and the same way every time.

## Your task

You maintain a rolling maintenance job. It drains hosts one at a time, and it
must never stop halfway through a host.

### 1. Stop between hosts

`Drain` already loops over the hosts, calls `drain(host)` and logs
`drained <host>` after each one. Add a check at the **start** of each loop
step: if `ctx.Err() != nil`, write this line to `log` and return the number of
hosts drained so far together with `ctx.Err()`:

```text
stopping: <N> hosts not drained
```

`N` is the number of hosts not yet drained. The host in progress always
finishes, because the check only happens before a host starts. Example, where
the context is cancelled while `web-2` is being drained:

```go
n, err := Drain(ctx, []string{"web-1", "web-2", "web-3", "web-4"}, drain, log)
// n == 2, errors.Is(err, context.Canceled) == true
// log:
// drained web-1
// drained web-2
// stopping: 2 hosts not drained
```

If the context is already cancelled before the first host, nothing is drained
and the log is just `stopping: 3 hosts not drained` (for 3 hosts).

### 2. Pick the exit code

Make `ExitCode` return:

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

Use `errors.Is`, as in *Errors*. You need to import `errors`.

### 3. Watch for signals

Write `WatchSignals`. Use `for sig := range sigs` to handle each signal:

- On the **first** signal, write
  `received <sig>: finishing current host (send again to force quit)` to
  `log`, then call `cancel()`.
- On the **second** signal, write `received <sig> again: forcing exit` to
  `log`, call `force()`, and return.
- If the channel closes, return.

Each log line ends with a newline. Print the signal with `%v`:

```text
received interrupt: finishing current host (send again to force quit)
received interrupt again: forcing exit
```

For `syscall.SIGTERM` the first line reads
`received terminated: finishing current host (send again to force quit)`.
With no signals at all, nothing is logged and neither function is called.

## Learn more

- [os/signal package](https://pkg.go.dev/os/signal)
- [Go by Example: Signals](https://gobyexample.com/signals)
- [Kubernetes: Termination of Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
- [context package](https://pkg.go.dev/context)
