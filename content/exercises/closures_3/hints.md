## Task 1: loop over attempts, keep a wait

Keep `wait := baseSeconds` before the loop. In the loop, return `true` if
`op(attempt)` succeeds. If it fails and this was not the last attempt, call
`sleep(wait)` and then double `wait`.

## Task 1: almost the answer

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

## Task 2: one shared counter

Declare `failures := 0` in `NewCircuitBreaker`. `record` adds 1 on a failure
and sets it back to 0 on a success. `isOpen` returns
`failures >= threshold`.
