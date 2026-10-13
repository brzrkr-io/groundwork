## Task 1: Build the Store

Both methods lock the store and then do one ordinary map operation on
`s.responses`.

### Closer

Start each method with `s.mu.Lock()` and `defer s.mu.Unlock()`. The two-value
lookup `resp, ok := s.responses[key]` gives you both return values of `Get`
at once.

### Answer

```go
// Get
s.mu.Lock()
defer s.mu.Unlock()
resp, ok := s.responses[key]
return resp, ok

// Put
s.mu.Lock()
defer s.mu.Unlock()
s.responses[key] = resp
```

## Task 2: Replay repeated requests

After the "no key" check, look the key up first and replay if it's there.
If not, the starter already records and sends the response; it just never
saves it.

### Closer

Set the header with `w.Header().Set("Idempotent-Replayed", "true")` before
calling `writeResponse`, because headers must be set before the status is
written. Add `store.Put(key, resp)` between `record` and `writeResponse`.

### Answer

```go
if resp, ok := store.Get(key); ok {
	w.Header().Set("Idempotent-Replayed", "true")
	writeResponse(w, resp)
	return
}
resp := record(next, r)
store.Put(key, resp)
writeResponse(w, resp)
```

## Task 3: Reject duplicates while the first is running

`Begin` checks both maps and claims the key, all under one lock. `Put` clears
the in-progress mark, and `Middleware` calls `Begin` before running `next`.

### Closer

Reading a missing key from a `map[string]bool` gives `false`, so
`s.inProgress[key]` works for keys that were never added. Use
`_, done := s.responses[key]` for the other check, and
`delete(s.inProgress, key)` in `Put`.

### Answer

```go
// Begin
s.mu.Lock()
defer s.mu.Unlock()
_, done := s.responses[key]
if done || s.inProgress[key] {
	return false
}
s.inProgress[key] = true
return true

// Put, after storing
delete(s.inProgress, key)

// Middleware, after the Get check
if !store.Begin(key) {
	http.Error(w, "request with this Idempotency-Key is still in progress", http.StatusConflict)
	return
}
```

`Begin` also refuses a key that finished just after `Get`, so `next` never
runs twice for one key.
