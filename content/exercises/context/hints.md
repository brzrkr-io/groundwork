## Task 1: Sleep, but wake up when cancelled

`time.Sleep` can't be interrupted. You need to wait on two things at once:
the timer and the context. Which statement waits on several channels?

### Closer

Use a `select` with `<-time.After(d)` and `<-ctx.Done()` as its two cases.
Each case returns something different.

### Answer

```go
select {
case <-time.After(d):
	return nil
case <-ctx.Done():
	return ctx.Err()
}
```

Whichever channel is ready first wins the `select`.

## Task 2: Stop a worker loop

A plain `for job := range jobs` can't notice cancellation while it waits for
the next job. Loop forever around a `select` that waits on both `jobs` and
`ctx.Done()`.

### Closer

Inside `for { select { ... } }`, receive with `job, ok := <-jobs`; `ok` is
`false` once `jobs` is closed. The other case is `<-ctx.Done()`.

### Answer

```go
for {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case job, ok := <-jobs:
		if !ok {
			return nil // jobs is closed
		}
		handle(job)
	}
}
```

## Task 3: Give a check a deadline

`context.Background()` never times out. Make a child context from it that
does, and always release it when you're done.

### Closer

`context.WithTimeout(parent, timeout)` returns `ctx, cancel`. Call
`defer cancel()` right after, then pass `ctx` to `check`.

### Answer

```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
return check(ctx)
```

`defer cancel()` frees the timer even if `check` returns early.
