## Task 1: lock, then use the map

Start both methods with `s.mu.Lock()` and `defer s.mu.Unlock()`. Then they
are ordinary map operations. The two-value lookup `resp, ok := s.responses[key]`
gives you both return values of `Get` at once.

## Task 2: three steps in order

After the "no key" check, the middleware does three things: look the key up
(replay and `return` if found), run `next` with `record`, then `Put` and
`writeResponse`. The starter already does the second and third, except for
`Put`. Set `Idempotent-Replayed` with `w.Header().Set(...)` before
calling `writeResponse`, because headers must be set before the status is
written.

## Task 3: Begin under one lock

```go
_, done := s.responses[key]
if done || s.inProgress[key] {
	return false
}
s.inProgress[key] = true
return true
```

Reading a missing key from a `map[string]bool` gives `false`, so
`s.inProgress[key]` works even for keys that were never added.

## Almost the answer for Middleware

```go
if resp, ok := store.Get(key); ok {
	w.Header().Set("Idempotent-Replayed", "true")
	writeResponse(w, resp)
	return
}
if !store.Begin(key) {
	http.Error(w, "request with this Idempotency-Key is still in progress", http.StatusConflict)
	return
}
resp := record(next, r)
store.Put(key, resp)
writeResponse(w, resp)
```

Don't forget `delete(s.inProgress, key)` in `Put`.
