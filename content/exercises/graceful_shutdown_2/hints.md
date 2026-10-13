## Task 1: Treat ErrServerClosed as success

`srv.Serve` never returns `nil`; after a normal stop it returns `http.ErrServerClosed`. Only other errors are real failures.

### Closer

Flip the condition to "the error is not `http.ErrServerClosed`": `!errors.Is(err, http.ErrServerClosed)`. Add `"errors"` to the imports.

### Answer

```go
if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
	return fmt.Errorf("serve: %w", err)
}
return nil
```

## Task 2: Give Shutdown a deadline

`context.Background()` never ends, so `Shutdown` can wait forever. Build a context that ends after `grace` and pass that instead.

### Closer

`context.WithTimeout(context.Background(), grace)` returns the context and a `cancel` function; `defer cancel()`.

### Answer

```go
ctx, cancel := context.WithTimeout(context.Background(), grace)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
	srv.Close()
	return fmt.Errorf("stop http: %w", err)
}
return nil
```

## Task 3: Shut down in the right order

HTTP must stop first so nothing new lands in the buffer. Then tell the worker to stop, and wait for it to finish its last upload.

### Closer

Save the HTTP error in a variable (`err := s.stopHTTP()`) so you can return it at the end, after `s.stopWorker()` and `s.worker.Wait()`.

### Answer

```go
err := s.stopHTTP()
s.stopWorker()
s.worker.Wait()
return err
```
