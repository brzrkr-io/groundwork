## Task 1: refill, then take

In `NewBucket`, add `tokens: float64(burst)` and `last: now()` to the struct
literal. In `Allow`, copy the refill code from the README, then check
`b.tokens < 1`.

## Task 1: almost the answer

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

## Task 2: a middleware

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

## Task 3: find or create

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
