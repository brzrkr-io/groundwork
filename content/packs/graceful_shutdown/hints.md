## Task 1: Set the server timeouts

`NewServer` returns an `http.Server` struct literal. Each timeout in the table is a field of that struct.

### Closer

Add one `Field: value,` line per timeout, for example `ReadHeaderTimeout: 5 * time.Second,`, next to `Handler: h`.

### Answer

```go
return &http.Server{
	Handler:           h,
	ReadHeaderTimeout: 5 * time.Second,
	ReadTimeout:       10 * time.Second,
	WriteTimeout:      30 * time.Second,
	IdleTimeout:       60 * time.Second,
}
```

## Task 2: Run until cancelled, then shut down

`Serve` blocks, so it can't run on the goroutine that waits for `ctx`. You need to wait for whichever happens first: `Serve` returning, or `ctx` being cancelled.

### Closer

Send `Serve`'s result on a buffered channel, `make(chan error, 1)`, from a goroutine. Then `select` on that channel and `ctx.Done()`.

### Answer

```go
serveErr := make(chan error, 1)
go func() {
	serveErr <- srv.Serve(ln)
}()

select {
case err := <-serveErr:
	return err
case <-ctx.Done():
	return Stop(srv, grace)
}
```

## Task 3: Force-close when the grace period runs out

`Shutdown` returns an error when its context ends before the requests finish. In that case, cut off what's left and report the error.

### Closer

Check `if err := srv.Shutdown(ctx); err != nil { ... }`. Inside, call `srv.Close()` and wrap with `fmt.Errorf("forced shutdown: %w", err)`. Add `"fmt"` to the imports.

### Answer

```go
if err := srv.Shutdown(ctx); err != nil {
	srv.Close()
	return fmt.Errorf("forced shutdown: %w", err)
}
return nil
```

`%w` keeps the original error, so `errors.Is(err, context.DeadlineExceeded)` still works.
