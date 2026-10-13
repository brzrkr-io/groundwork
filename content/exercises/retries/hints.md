## Task 1: double in a loop

Start with `d := base` and double it once per attempt with a
`for range attempt` loop. Each time round, check whether `d` has reached
`limit`; if it has, return `limit`.

## Task 2: half plus a random part

Work out `half := d / 2` first. `rnd()` is a `float64`, so multiply it by
`float64(half)` and convert the result back with `time.Duration(...)`. Add
that to `half`.

## Task 3: the order of the checks

After `err = op(ctx)`, check in this order: success or permanent error
(return), last attempt (`attempt == p.MaxAttempts-1`, so `break`), then sleep.

## Task 3: almost the answer

```go
for attempt := range p.MaxAttempts {
	err = op(ctx)
	if err == nil || !errors.Is(err, ErrTemporary) {
		return err
	}
	if attempt == p.MaxAttempts-1 {
		break
	}
	wait := Jitter(Backoff(attempt, p.Base, p.Max), p.Rand)
	if sleepErr := p.Sleep(ctx, wait); sleepErr != nil {
		return sleepErr
	}
}
```
