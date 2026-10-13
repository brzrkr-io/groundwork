## Task 1: change the condition

`srv.Serve` never returns `nil`. The condition should be "the error is *not*
`http.ErrServerClosed`":

```go
if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
```

## Task 2: a timeout context

Create the context before calling `Shutdown`, and release it when you're done:

```go
ctx, cancel := context.WithTimeout(context.Background(), grace)
defer cancel()
```

Then pass `ctx` to `srv.Shutdown`.

## Task 3: almost the answer

```go
err := s.stopHTTP()
s.stopWorker()
s.worker.Wait()
return err
```
