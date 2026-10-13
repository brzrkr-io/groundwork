# Graceful shutdown: HTTP server and worker

Real services do more than answer HTTP. A background goroutine flushes
buffers or uploads batches. On shutdown the parts must stop in the right
order, or the last batch of data is lost. Here you fix three classic shutdown
bugs.

**New here:** `errors.Is` with `http.ErrServerClosed`, stopping parts in
order.

## Learn

### Recap

As in *Graceful shutdown: drain in-flight requests*: `srv.Serve(ln)` blocks
until the server stops. `srv.Shutdown(ctx)` refuses new connections and waits
for in-flight requests until `ctx` ends. `srv.Close()` cuts off what's left.
After either, `Serve` returns `http.ErrServerClosed`, meaning "stopped
normally".

### Checking for one specific error

`errors.Is(err, target)` reports whether `err` is `target`, or wraps it. It
returns `false` for `nil`:

```go
if errors.Is(err, os.ErrNotExist) {
	return nil // a missing file is fine here
}
```

> **Why is it written like this?**
> - **Go requires** the `"errors"` import.
> - **Convention:** compare errors with `errors.Is`, not `==`, so wrapped
>   errors still match.

### `context.Background()` never ends

`context.Background()` has no deadline. Passing it to `Shutdown` means "wait
as long as it takes". One stuck request and the process never exits;
Kubernetes `SIGKILL`s it and your clean-up never runs. Wrap it with
`context.WithTimeout` (and `defer cancel()`) to get a deadline.

### Stopping in the right order

`metrics-relay` handlers put metrics in a buffer. A background *worker*
uploads the buffer in batches, and one last batch when its context is
cancelled. The safe order:

1. **Stop HTTP first.** After that no handler runs, so nothing new lands in
   the buffer.
2. **Then cancel the worker.** It uploads the last batch.
3. **Then wait for the worker.** If you return first, the process exits
   mid-upload.

`Service` stores the pieces: `stopHTTP` (a function, so tests can pass a
fake), `stopWorker` (the cancel function) and `worker` (a `sync.WaitGroup`).

### The order test

`TestShutdownOrder` records events and checks their order inside
`synctest.Test`, where the clock is fake. The "1 second" HTTP drain and "2
second" upload take no real time, and the order is always the same.

## Your task

The `metrics-relay` service loses data on every deploy and sometimes never
exits. Fix its shutdown code.

### 1. Treat ErrServerClosed as success

**Goal:** make `Serve` return `nil` when `srv.Serve` returns
`http.ErrServerClosed`.

Check with `errors.Is` (add `"errors"` to the imports). Any other error is
still returned as `serve: <error>`.

```text
Serve(srv, ln), then srv.Shutdown(...)  => nil
Serve(srv, ln), then srv.Close()        => nil
Serve(srv, closedListener)              => "serve: accept tcp ...: use of closed network connection"
```

**Think it through:**
- The `if` currently asks "is there any error?". What question should it ask
  instead?
- `srv.Serve` always returns a non-nil error, so you only need one condition.

### 2. Give Shutdown a deadline

**Goal:** make `StopHTTP` give up after `grace` instead of waiting forever.

Pass `Shutdown` a context that expires after `grace`. The rest (`srv.Close()`
and the `stop http:` wrapping) is already there.

```text
StopHTTP(srv, 20*time.Millisecond)   // one request stuck forever
=> "stop http: context deadline exceeded" after about 20ms
   errors.Is(err, context.DeadlineExceeded) is true

StopHTTP(srv, time.Second)           // no requests
=> nil
```

**Think it through:**
- Which argument to `Shutdown` is the problem?
- Build the new context *before* the `if` line, from `context.Background()`
  and `grace`.

### 3. Shut down in the right order

**Goal:** make `Service.Shutdown` stop HTTP, then the worker, wait for the
worker, and return the HTTP error.

1. Call `s.stopHTTP()` and keep its error.
2. Call `s.stopWorker()`.
3. Call `s.worker.Wait()`.
4. Return the error from step 1.

```text
Shutdown() with a 1s HTTP drain and a 2s final upload
=> events: "http stopped", "worker stopping", "worker flushed"
   (all three recorded by the time Shutdown returns)
```

**Think it through:**
- `return s.stopHTTP()` ends the function at once. How do you keep its error
  for later?
- `s.stopWorker` and `s.stopHTTP` are fields that hold functions: add `()`
  to call them.

## Learn more

- [pkg.go.dev: http.Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [pkg.go.dev: errors.Is](https://pkg.go.dev/errors#Is)
- [Go blog: Testing concurrent code with testing/synctest](https://go.dev/blog/synctest)
- [Kubernetes: Container lifecycle hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/)
