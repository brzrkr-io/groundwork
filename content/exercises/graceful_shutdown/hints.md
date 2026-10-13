## Task 1: struct fields

Add the four fields to the `&http.Server{...}` literal, for example
`ReadHeaderTimeout: 5 * time.Second,`.

## Task 2: a goroutine and a select

`Serve` blocks, so it can't run on the same goroutine that waits for `ctx`.
Send its result on a buffered channel, then `select` on that channel and
`ctx.Done()`.

## Task 2: almost the answer

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

## Task 3: the fallback

```go
if err := srv.Shutdown(ctx); err != nil {
	srv.Close()
	return fmt.Errorf("forced shutdown: %w", err)
}
return nil
```
