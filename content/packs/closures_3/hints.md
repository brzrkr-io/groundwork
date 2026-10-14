## Task 1: Retry with backoff

Loop over the attempts like in *Functions as values*, but also keep a `wait`
number that doubles after each sleep. Only sleep when an attempt failed and
another attempt is still coming.

### Closer

Start with `wait := baseSeconds` before the loop. In the loop, return `true`
if `op(attempt)` succeeds. Otherwise, `if attempt < attempts`, call
`sleep(wait)` and then `wait *= 2`. After the loop, return `false`.

### Answer

```go
wait := baseSeconds
for attempt := 1; attempt <= attempts; attempt++ {
	if op(attempt) {
		return true
	}
	if attempt < attempts {
		sleep(wait)
		wait *= 2
	}
}
return false
```

## Task 2: Build a circuit breaker

Both functions need to see the same number: how many failures in a row there
have been. Declare it once in `NewCircuitBreaker` so both closures share it.

### Closer

Declare `failures := 0` before building the two functions. In `record`, set
`failures = 0` when `ok` is true, and do `failures++` when it is false.
`isOpen` returns `failures >= threshold`.

### Answer

```go
failures := 0
record = func(ok bool) {
	if ok {
		failures = 0
	} else {
		failures++
	}
}
isOpen = func() bool {
	return failures >= threshold
}
return record, isOpen
```
