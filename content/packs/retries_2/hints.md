## Task 1: Fix the retry loop

`attempt` starts at `0`. Count how many values the loop goes through with
`<=`, then think about what should happen after the very last call, before
`Sleep` runs.

### Closer

Use `<` instead of `<=` in the loop condition. Then, after checking the
error and before sleeping, stop early on the last attempt:
`if attempt == p.MaxAttempts-1 { ... }`.

### Answer

```go
var err error
for attempt := 0; attempt < p.MaxAttempts; attempt++ {
	err = op(ctx)
	if err == nil || !p.ShouldRetry(err) {
		return err
	}
	if attempt == p.MaxAttempts-1 {
		break
	}
	if sleepErr := p.Sleep(ctx, Backoff(attempt, p.Base, p.Max)); sleepErr != nil {
		return sleepErr
	}
}
return err
```

The loop now runs exactly `MaxAttempts` times, and `break` skips the sleep
after the final call.

## Task 2: Classify retryable errors

`Retryable` only needs to pick out four codes. For `IsRetryable`, look for
the `Retryable` interface anywhere in the error chain, not for `*StatusError`.

### Closer

A `case` can list several values: `case 429, 502, 503, 504:`. For the chain,
declare `var r Retryable` and pass `&r` to `errors.As`; it returns `true`
when it found one. Add `"errors"` to the imports.

### Answer

```go
func (e *StatusError) Retryable() bool {
	switch e.Code {
	case 429, 502, 503, 504:
		return true
	}
	return false
}

func IsRetryable(err error) bool {
	var r Retryable
	return errors.As(err, &r) && r.Retryable()
}
```

`errors.As` fills `r` with the first wrapped error that has a `Retryable`
method, and returns `false` for `nil`.

## Task 3: Sleep, but wake up on cancel

A `select` waits until one of its channels is ready. You need two cases: one
for the time running out and one for the context being cancelled.

### Closer

`time.After(d)` returns a channel that receives once `d` has passed.
`ctx.Done()` is a channel that closes on cancel. Receive from each with
`case <-...:`.

### Answer

```go
select {
case <-time.After(d):
	return nil
case <-ctx.Done():
	return ctx.Err()
}
```
