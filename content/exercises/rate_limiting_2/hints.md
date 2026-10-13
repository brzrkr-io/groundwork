## Task 1: two missing pieces

Wrap the new total in `min(b.burst, ...)`, and after adding the tokens set
`b.last = now`.

## Task 2: the arithmetic

```go
if b.tokens >= 1 {
	return 0
}
missing := 1 - b.tokens
return time.Duration(missing / b.rate * float64(time.Second))
```

Don't forget `b.mu.Lock()`, `defer b.mu.Unlock()` and `b.refill()` first.

## Task 2: the header

```go
secs := math.Ceil(b.RetryAfter().Seconds())
w.Header().Set("Retry-After", strconv.Itoa(int(secs)))
```

## Task 3: delete while ranging

```go
l.mu.Lock()
defer l.mu.Unlock()
removed := 0
now := l.now()
for ip, c := range l.clients {
	if now.Sub(c.seen) >= idle {
		delete(l.clients, ip)
		removed++
	}
}
return removed
```
