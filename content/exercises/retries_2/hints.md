## Task 1: count the calls

`attempt` starts at `0`. With `<=`, how many values does it take before the
loop stops? Then, before sleeping, check whether this was the last attempt
(`attempt == p.MaxAttempts-1`) and `break` if so.

## Task 2: switch and errors.As

```go
switch e.Code {
case 429, 502, 503, 504:
	return true
}
return false
```

For `IsRetryable`, declare `var r Retryable` and pass `&r` to `errors.As`.

## Task 3: a two-way select

`time.After(d)` returns a channel that receives once `d` has passed.
`ctx.Done()` is a channel that closes when the context is cancelled.

## Almost the answer

```go
func IsRetryable(err error) bool {
	var r Retryable
	return errors.As(err, &r) && r.Retryable()
}

func SleepCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
```
