# Graceful shutdown: drain in-flight requests

Every deploy stops your old pods. If a pod just dies, the requests it was
handling fail and users see errors on each release. A *graceful shutdown*
stops taking new work, finishes the work it has, then exits.

**New here:** `http.Server` timeouts, `Serve`, `Shutdown` and `Close`,
waiting on a goroutine or a context with `select`.

## Learn

### What happens during a rolling deploy

1. Kubernetes removes the old pod from the Service.
2. It sends the old container `SIGTERM`.
3. It waits up to `terminationGracePeriodSeconds` (30 seconds by default).
4. If the process still runs, it sends `SIGKILL`, which can't be caught.

Between `SIGTERM` and `SIGKILL`: stop accepting connections, let in-flight
requests finish, and exit.

### Server timeouts

`http.ListenAndServe` has no timeouts. A client can send its headers one byte
a minute and hold a connection forever; enough of them use up your server
(the *Slowloris* attack). A production `http.Server` sets these fields:

| Field | Limits |
|---|---|
| `ReadHeaderTimeout` | reading the request headers (the Slowloris fix) |
| `ReadTimeout` | reading the whole request, headers and body |
| `WriteTimeout` | from the end of the headers to the end of the response |
| `IdleTimeout` | how long a keep-alive connection waits for its next request |

> **Why is it written like this?**
> - **Go requires** the type `time.Duration` for each field, so write a
>   number times a unit, such as `3 * time.Minute`.
> - **Convention:** in a struct literal, one `Field: value,` per line, with a
>   trailing comma. gofmt lines up the values.

### Serve, Shutdown and Close

`srv.Serve(ln)` blocks until the server stops. Two methods stop it:

| Method | What it does |
|---|---|
| `srv.Shutdown(ctx)` | Refuses new connections, closes idle ones, waits for active requests. Returns `nil` when they're done, or `ctx.Err()` if `ctx` ends first. |
| `srv.Close()` | Closes everything right now. In-flight requests are cut off. |

After either one, `Serve` returns `http.ErrServerClosed`: the normal way to
stop, not a failure.

Give `Shutdown` a deadline with `context.WithTimeout`, built from
`context.Background()`. Don't build it from the context that told you to stop:
that one is already cancelled.

### Waiting for whichever comes first

`select` waits on several channels and runs the case that is ready first. A
blocking call runs in a goroutine that sends its result on a channel:

```go
done := make(chan int, 1) // buffer of 1: the send never blocks
go func() { done <- countLines(path) }()

select {
case n := <-done:
	fmt.Println("lines:", n)
case <-ctx.Done():
	fmt.Println("cancelled")
}
```

The buffer matters. If `ctx` wins, nobody reads `done`; with a buffer of 1 the
goroutine can still send and exit instead of leaking.

> **Why is it written like this?**
> - **Go requires** `go` in front of a *call*, so the anonymous `func() {...}`
>   ends with `()`.
> - **Convention:** `<-ctx.Done()` is how you wait for cancellation.

### Wrapping an error

As in *Wrapping errors*, `%w` in `fmt.Errorf` adds context but keeps the
original error findable with `errors.Is`:

```go
return fmt.Errorf("load config: %w", err)
```

### Wiring it to signals

In `main`, `signal.NotifyContext` (the one-call version of the watcher from
*Signals: graceful stop for a rolling drain*) gives a context that is
cancelled on Ctrl-C or `SIGTERM`, which is what `Run` waits on:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
err = shutdown.Run(ctx, shutdown.NewServer(mux), ln, 20*time.Second)
```

Keep the grace period shorter than `terminationGracePeriodSeconds`.

## Your task

The `deploy-api` service drops requests on every rollout. Give it timeouts and
a graceful shutdown. The tests run a real server on `127.0.0.1` with a handler
that stays busy until the test lets it reply.

### 1. Set the server timeouts

**Goal:** make `NewServer` return an `*http.Server` with `Handler: h` and
these four timeouts.

| Field | Value |
|---|---|
| `ReadHeaderTimeout` | 5 seconds |
| `ReadTimeout` | 10 seconds |
| `WriteTimeout` | 30 seconds |
| `IdleTimeout` | 60 seconds |

**Think it through:**
- The struct literal already has `Handler: h,`. Add the other fields next to it.
- How do you write "5 seconds" as a `time.Duration`?

### 2. Run until cancelled, then shut down

**Goal:** make `Run` serve in the background and return when either serving
fails or `ctx` is cancelled.

- Start `srv.Serve(ln)` in a goroutine that sends its error on a buffered
  channel.
- If `Serve` returns first, return its error. (That only happens when serving
  failed, for example on a closed listener.)
- If `ctx` is cancelled first, return `Stop(srv, grace)`.

```text
ctx cancelled while GET /deploy is in flight
=> new connections are refused
=> GET /deploy still gets 200 "done"
=> Run returns nil
```

**Think it through:**
- What type of value does `Serve` return? That's the channel's element type.
- Which two things does the `select` wait on?

### 3. Force-close when the grace period runs out

**Goal:** in `Stop`, if `Shutdown` fails, close the server and return the
error wrapped as `forced shutdown: <error>`.

`Stop` already calls `Shutdown` with a `grace` timeout. On error, call
`srv.Close()`, then return the error wrapped with `%w`. On success return
`nil`. Add `"fmt"` to the imports.

```go
Stop(srv, 20*time.Millisecond) // a request is stuck forever
// => error "forced shutdown: context deadline exceeded"
//    errors.Is(err, context.DeadlineExceeded) is true
//    the stuck request is cut off

Stop(srv, time.Second) // nothing in flight
// => nil
```

**Think it through:**
- Right now `Shutdown`'s result is returned directly. Store it in `err` and
  check it with `if`.
- The fixed text is `forced shutdown: ` (with a space after the colon); the
  rest comes from `err`.

## Learn more

- [pkg.go.dev: http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [pkg.go.dev: signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)
- [Kubernetes: Termination of Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
- [Go by Example: Signals](https://gobyexample.com/signals)
