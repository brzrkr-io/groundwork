## Task 1: Fix the refill

`refill` is missing two things. One caps the total at `burst`; the other
remembers when this refill happened.

### Closer

Wrap the new total in `min(b.burst, ...)` instead of using `+=`, and after
adding tokens set `b.last` to the current time.

### Answer

```go
b.tokens = min(b.burst, b.tokens+now.Sub(b.last).Seconds()*b.rate)
b.last = now
```

## Task 2: Report an accurate Retry-After

`RetryAfter` works out the missing part of a token and divides by the rate.
Then the middleware turns that duration into whole seconds, rounded up.

### Closer

Start with `b.mu.Lock()`, `defer b.mu.Unlock()` and `b.refill()`. Convert
seconds to a duration with `time.Duration(secs * float64(time.Second))`. In
`Middleware`, use `math.Ceil(d.Seconds())` and `strconv.Itoa(int(...))`, and
add `"math"` and `"strconv"` to the imports.

### Answer

```go
// RetryAfter
b.mu.Lock()
defer b.mu.Unlock()
b.refill()
if b.tokens >= 1 {
	return 0
}
missing := 1 - b.tokens
return time.Duration(missing / b.rate * float64(time.Second))

// Middleware
secs := math.Ceil(b.RetryAfter().Seconds())
w.Header().Set("Retry-After", strconv.Itoa(int(secs)))
```

## Task 3: Evict idle clients

Range over the clients map with the lock held, and delete those whose last
request was long enough ago. Keep a count of how many you remove.

### Closer

Deleting while ranging is fine: `delete(l.clients, ip)`. A client is idle
when `now.Sub(c.seen) >= idle`, with `now := l.now()`.

### Answer

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
