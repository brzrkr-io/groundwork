## Task 1: Compute the backoff

Start from `base` and double it once per attempt. Stop as soon as the value reaches `limit`.

### Closer

Use `d := base` and a `for range attempt` loop with `d *= 2`. Inside the loop, `if d >= limit { return limit }`.

### Answer

```go
d := base
for range attempt {
	d *= 2
	if d >= limit {
		return limit
	}
}
return d
```

## Task 2: Add jitter

The result is half of `d` plus a random fraction of the other half. Call `rnd()` only once.

### Closer

Work out `half := d / 2`. `rnd()` is a `float64`, so compute `rnd() * float64(half)` and convert back with `time.Duration(...)`.

### Answer

```go
half := d / 2
return half + time.Duration(rnd()*float64(half))
```

## Task 3: Retry the operation

After each call to `op`, decide in order: return (success or permanent error), stop (last attempt), or sleep and go again.

### Closer

Loop with `for attempt := range p.MaxAttempts`. Success or permanent is `err == nil || !errors.Is(err, ErrTemporary)`; the last attempt is `attempt == p.MaxAttempts-1`. The wait is `Jitter(Backoff(attempt, p.Base, p.Max), p.Rand)`.

### Answer

```go
var err error
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
return err
```

Breaking before the sleep means there's no pointless wait after the final attempt.
