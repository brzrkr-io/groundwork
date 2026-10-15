## Task 1: Build the bucket

A new bucket starts full, so `NewBucket` needs two more fields. In `Allow`,
refill first using the time since `last`, then decide whether there's a whole
token to take.

### Closer

In the struct literal add `tokens: float64(burst)` and `last: now()`. To
refill, use `b.tokens = min(b.burst, b.tokens+elapsed*b.rate)`, then
`b.last = now`. After that, check `b.tokens < 1`.

### Answer

```go
now := b.now()
b.tokens = min(b.burst, b.tokens+now.Sub(b.last).Seconds()*b.rate)
b.last = now
if b.tokens < 1 {
	return false
}
b.tokens--
return true
```

## Task 2: Reject requests with 429

Return a new handler that asks the bucket first. If the answer is no, reply
with a 429 and stop; otherwise hand the request to `next`.

### Closer

Wrap a function with `http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ... })`.
Set the header before writing the status:
`w.Header().Set("Retry-After", "1")`, then `http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)`.

### Answer

```go
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if !b.Allow() {
		w.Header().Set("Retry-After", "1")
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	next.ServeHTTP(w, r)
})
```

## Task 3: Give every client its own bucket

Key the map by IP only, not the port. Look the bucket up under the lock, and
create and store one if it's missing.

### Closer

`ip, _, _ := net.SplitHostPort(r.RemoteAddr)` (add `"net"` to the imports).
Use the comma-ok form `b, ok := p.buckets[ip]` between `p.mu.Lock()` and
`p.mu.Unlock()`.

### Answer

```go
ip, _, _ := net.SplitHostPort(r.RemoteAddr)
p.mu.Lock()
b, ok := p.buckets[ip]
if !ok {
	b = NewBucket(p.rate, p.burst, p.now)
	p.buckets[ip] = b
}
p.mu.Unlock()
return b.Allow()
```

The bucket has its own lock, so `p.mu` can be released before calling
`Allow`.
