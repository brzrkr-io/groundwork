## Task 1: select on two channels

Replace `time.Sleep(d)` with a `select` that has two cases: `<-time.After(d)`
(return `nil`) and `<-ctx.Done()` (return `ctx.Err()`).

## Task 2: a loop around a select

Use `for { select { ... } }` with two cases:

```go
case <-ctx.Done():
	return ctx.Err()
case job, ok := <-jobs:
	if !ok {
		return nil // jobs is closed
	}
	handle(job)
```

A plain `for job := range jobs` can't notice cancellation while it is
waiting for the next job.

## Task 3: WithTimeout and defer cancel

```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
return check(ctx)
```
