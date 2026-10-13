## Task 1: Name the states

Each of the three constants gets its own name. Anything else falls through to
`"unknown"`.

### Closer

Use `switch s { case Closed: return "closed" ... }` with one `case` per
state, and put `return "unknown"` after the switch.

### Answer

```go
switch s {
case Closed:
	return "closed"
case Open:
	return "open"
case HalfOpen:
	return "half-open"
}
return "unknown"
```

## Task 2: Count failures and fail fast

In `record`, the success case already returns early, so any code after it
handles a failure. In `allow`, an open breaker just says no.

### Closer

In `record`: `b.failures++`, then compare it with `b.threshold` and, when
reached, set `b.state` and `b.openedAt = b.now()`. In `allow`:
`if b.state == Open { return ErrOpen }`.

### Answer

```go
// allow
if b.state == Open {
	return ErrOpen
}
return nil

// record, after the success branch
b.failures++
if b.failures >= b.threshold {
	b.state = Open
	b.openedAt = b.now()
}
```

## Task 3: Let a trial call through

`allow` now has two non-closed states to handle: `Open` checks whether the
cooldown is over, and `HalfOpen` means a trial is already running. `record`
needs to close on success and reopen on a failed trial.

### Closer

Use `switch b.state` with `case Open:` and `case HalfOpen:`. The cooldown
check is `b.now().Sub(b.openedAt) < b.cooldown`. In `record`, add
`b.state = Closed` to the success branch and `b.state == HalfOpen ||` to the
failure condition.

### Answer

```go
// allow
switch b.state {
case Open:
	if b.now().Sub(b.openedAt) < b.cooldown {
		return ErrOpen
	}
	b.state = HalfOpen
case HalfOpen:
	return ErrOpen
}
return nil

// record
if err == nil {
	b.state = Closed
	b.failures = 0
	return
}
b.failures++
if b.state == HalfOpen || b.failures >= b.threshold {
	b.state = Open
	b.openedAt = b.now()
}
```

A failed trial reopens straight away and resets `openedAt`, so the cooldown
starts over.
