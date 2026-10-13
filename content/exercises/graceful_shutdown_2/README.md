# Graceful shutdown: HTTP server and worker

Real services do more than answer HTTP. A background goroutine flushes
buffers, consumes a queue or uploads batches. On shutdown, the parts have to
stop in the right order, or the last batch of data is lost. This exercise has
three classic shutdown bugs to fix.

## Learn

### Recap

As in *Graceful shutdown: drain in-flight requests*, `srv.Serve(ln)` blocks
until the server stops. `srv.Shutdown(ctx)` closes the listener and waits for
in-flight requests until `ctx` ends, and `srv.Close()` cuts off whatever is
left. After either one, `Serve` returns `http.ErrServerClosed`, which means
"stopped normally". Give `Shutdown` a deadline built from
`context.Background()` with `context.WithTimeout`.

### `context.Background()` never ends

`context.Background()` has no deadline and is never cancelled. Passing it to
`Shutdown` means "wait as long as it takes". One stuck request, such as a
client streaming a huge upload, and the process never exits. Kubernetes
`SIGKILL`s it at the end of the grace period, and all the clean-up you planned
never runs.

### Stopping in the right order

The `metrics-relay` service receives metrics over HTTP. Its handlers put them
in a buffer. A background *worker* uploads the buffer in batches and uploads
one last batch when its context is cancelled. Stopping order matters:

1. **Stop HTTP first.** After `Shutdown` returns, no handler is running, so
   nothing new can land in the buffer.
2. **Then cancel the worker's context.** It uploads the last batch.
3. **Then wait for the worker** with its `sync.WaitGroup`. If `main` returns
   first, the process exits and the upload is cut off halfway.

Cancel the worker first, and requests still in flight add metrics to a buffer
nobody will upload.

`Service` keeps the pieces it needs to stop. `stopHTTP` is a function, so
production code passes a closure around `StopHTTP` and the test passes a fake:

```go
svc := shutdown.Start(
	func() error { return shutdown.StopHTTP(srv, 10*time.Second) },
	flusher.Run, // func(ctx context.Context): flushes and returns when ctx ends
)
<-ctx.Done() // from signal.NotifyContext
err := svc.Shutdown()
```

### The order test

`TestShutdownOrder` records events in a slice guarded by a mutex and checks
their order. It runs inside `synctest.Test`, where the clock is fake. Sleeps
end only when every goroutine is waiting, and then they end instantly. So the
"1 second" HTTP drain and the "2 second" upload take no real time, and the
order of events is always the same.

## Your task

The `metrics-relay` service loses data on every deploy and sometimes never
exits. Fix its shutdown code.

### 1. Treat ErrServerClosed as success

`Serve` reports every return from `srv.Serve` as a failure, including a normal
shutdown. Make it return `nil` when the error is `http.ErrServerClosed` (check
with `errors.Is`, and add `"errors"` to the imports). Any other error is still
returned as `serve: <error>`.

```text
Serve(srv, ln), then srv.Shutdown(...)  => nil
Serve(srv, ln), then srv.Close()        => nil
Serve(srv, closedListener)              => "serve: accept tcp ...: use of closed network connection"
```

### 2. Give Shutdown a deadline

`StopHTTP` passes `context.Background()` to `Shutdown`, so a stuck request
makes it wait forever. Pass a context that expires after `grace` instead.
The rest (`srv.Close()` and the `stop http:` wrapping) is already there.

```text
StopHTTP(srv, 20*time.Millisecond)   // one request stuck forever
=> "stop http: context deadline exceeded" after about 20ms
   errors.Is(err, context.DeadlineExceeded) is true

StopHTTP(srv, time.Second)           // no requests
=> nil
```

### 3. Shut down in the right order

`Service.Shutdown` cancels the worker first and never waits for it. Make it:

1. call `s.stopHTTP()` and keep its error,
2. call `s.stopWorker()`,
3. call `s.worker.Wait()`,
4. return the error from step 1.

```text
Shutdown() with a 1s HTTP drain and a 2s final upload
=> events: "http stopped", "worker stopping", "worker flushed"
   (all three recorded by the time Shutdown returns)
```

## Learn more

- [pkg.go.dev: http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [pkg.go.dev: context.Background](https://pkg.go.dev/context#Background)
- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
- [Kubernetes: Container lifecycle hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/)
