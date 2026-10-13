# Graceful shutdown: drain in-flight requests

Every deploy stops your old pods. If a pod just dies, every request it was
handling fails, and users see errors on each release. A *graceful shutdown*
stops taking new work, finishes the work it already has, and then exits.

## Learn

### What happens during a rolling deploy

1. Kubernetes starts a new pod and removes the old one from the Service.
2. It sends the old container `SIGTERM`.
3. It waits up to `terminationGracePeriodSeconds` (30 seconds by default).
4. If the process is still running, it sends `SIGKILL`, which can't be caught.

So you have a few seconds between `SIGTERM` and `SIGKILL`. In that time:
stop accepting new connections, let in-flight requests finish, and exit.

### Server timeouts

`http.ListenAndServe` uses a server with no timeouts. A client can open a
connection and send its headers one byte a minute, holding the connection open
forever. Enough of those clients use up your server. That attack is called
*Slowloris*. A production server sets timeouts on its `http.Server`:

| Field | Limits |
|---|---|
| `ReadHeaderTimeout` | reading the request headers (the Slowloris fix) |
| `ReadTimeout` | reading the whole request, headers and body |
| `WriteTimeout` | from the end of reading the headers to the end of writing the response |
| `IdleTimeout` | how long a keep-alive connection may wait for its next request |

### Serve, Shutdown and Close

```go
srv := &http.Server{Handler: mux}
ln, err := net.Listen("tcp", ":8080")
err = srv.Serve(ln) // blocks until the server stops
```

`Serve` blocks, so it runs in a goroutine while another goroutine waits for the
signal to stop. Two methods stop a server:

| Method | What it does |
|---|---|
| `srv.Shutdown(ctx)` | Closes the listener (new connections are refused), closes idle connections, then waits for active requests to finish. Returns `nil` when they are all done, or `ctx.Err()` if `ctx` ends first. |
| `srv.Close()` | Closes everything right now. In-flight requests are cut off. |

Once `Shutdown` or `Close` is called, `Serve` returns `http.ErrServerClosed`.
That is the normal way to stop, not a failure.

`Shutdown` waits as long as its context lets it, so give it a deadline. Build
that context from `context.Background()`, not from the context that told you
to stop: that one is already cancelled, and `Shutdown` would give up at once.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
	srv.Close() // grace period over: cut off whatever is left
}
```

### Waiting for a goroutine or a context

`Run` waits for whichever happens first: `Serve` fails, or `ctx` is cancelled.
A `select` over two channels does that. A buffered channel of size 1 lets the
`Serve` goroutine send its result and exit, even if nobody is reading anymore:

```go
serveErr := make(chan error, 1)
go func() { serveErr <- srv.Serve(ln) }()

select {
case err := <-serveErr:
	// Serve failed before anyone asked us to stop.
case <-ctx.Done():
	// Time to shut down.
}
```

### Wiring it to signals

In `main`, `signal.NotifyContext` gives you a context that is cancelled on
Ctrl-C or `SIGTERM`, which is exactly what `Run` needs:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
ln, err := net.Listen("tcp", ":8080")
// handle err
err = shutdown.Run(ctx, shutdown.NewServer(mux), ln, 20*time.Second)
```

Keep the grace period shorter than `terminationGracePeriodSeconds`, so you
finish before `SIGKILL` arrives.

## Your task

The `deploy-api` service drops requests on every rollout. Give it timeouts and
a graceful shutdown. The tests run a real server on `127.0.0.1` with a handler
that stays busy until the test lets it reply.

### 1. Set the server timeouts

Make `NewServer` return an `*http.Server` with `Handler: h` and these
timeouts:

| Field | Value |
|---|---|
| `ReadHeaderTimeout` | `5 * time.Second` |
| `ReadTimeout` | `10 * time.Second` |
| `WriteTimeout` | `30 * time.Second` |
| `IdleTimeout` | `60 * time.Second` |

### 2. Run until cancelled, then shut down

Rewrite `Run`:

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

### 3. Force-close when the grace period runs out

`Stop` already calls `Shutdown` with a `grace` timeout. If `Shutdown` returns
an error, call `srv.Close()` and return the error wrapped with `%w`:

```go
Stop(srv, 20*time.Millisecond) // a request is stuck forever
// => error "forced shutdown: context deadline exceeded"
//    errors.Is(err, context.DeadlineExceeded) is true
//    the stuck request is cut off

Stop(srv, time.Second) // nothing in flight
// => nil
```

Add `"fmt"` to the imports for `fmt.Errorf`.

## Learn more

- [pkg.go.dev: http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [pkg.go.dev: signal.NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)
- [Kubernetes: Termination of Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
- [Go by Example: Signals](https://gobyexample.com/signals)
