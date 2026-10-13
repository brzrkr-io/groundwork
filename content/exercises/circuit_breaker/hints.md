## Task 1: a switch

Use a `switch s { case Closed: ... }` with one `return` per state. Put
`return "unknown"` after the switch for every other value.

## Task 2: two small pieces

`allow` only needs an `if b.state == Open { return ErrOpen }`. In `record`,
the `err == nil` case already returns early, so after it you know the call
failed: `b.failures++`, then compare it with `b.threshold`.

## Task 3: the cooldown check

Inside the `Open` case of `allow`, compare `b.now().Sub(b.openedAt)` with
`b.cooldown`. A `switch b.state` with `case Open:` and `case HalfOpen:` keeps
it tidy. In `record`, the success branch also sets `b.state = Closed`, and the
failure branch opens the breaker when `b.state == HalfOpen` too.

## Near-solution

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
