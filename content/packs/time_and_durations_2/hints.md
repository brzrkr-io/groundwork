## Task 1: Fix the archive name

Go layouts aren't any date you like: they must spell out the one reference
time, `Mon Jan 2 15:04:05 MST 2006`. Which pieces of it does the name need?

### Closer

Year, month, day, hour and minute are `2006`, `01`, `02`, `15` and `04`.
Keep the `-` between the date part and the time part.

### Answer

```go
return host + "-" + at.UTC().Format("20060102-1504") + ".tar.gz"
```

Format replaces each reference piece with the matching part of `at`.

## Task 2: Fix the retry delay

`time.Duration(attempt * 30)` is 30 *nanoseconds* per attempt, because a
`Duration` counts nanoseconds. Multiply by a unit to get seconds.

### Closer

Convert `attempt` first, then multiply by the unit:
`time.Duration(attempt) * 30 * time.Second`. The cap check can stay.

### Answer

```go
delay := time.Duration(attempt) * 30 * time.Second
if delay > MaxRetryDelay {
	return MaxRetryDelay
}
return delay
```

Multiplying by `time.Second` turns the plain count into real seconds.

## Task 3: Find the next run

Round `now` down to the start of its interval, then step forward one
interval. That works even when `now` is already on a boundary.

### Closer

`now.Truncate(every)` rounds down to a multiple of `every`, and `.Add(every)`
moves forward by one interval.

### Answer

```go
func NextRun(now time.Time, every time.Duration) time.Time {
	return now.Truncate(every).Add(every)
}
```

On a boundary, `Truncate` changes nothing, so the result is a full interval
later, as required.
