## Task 1: Rotate through the backends

Lock first and defer the unlock. Then remember the backend at `rr.next`,
move `rr.next` on (wrapping with `%`), and return the remembered address.

### Closer

`b := rr.backends[rr.next]` grabs the backend before you move the turn on.
Wrap with `(rr.next + 1) % len(rr.backends)`.

### Answer

```go
rr.mu.Lock()
defer rr.mu.Unlock()
b := rr.backends[rr.next]
rr.next = (rr.next + 1) % len(rr.backends)
return b.addr, nil
```

## Task 2: Skip backends that are down

`MarkDown` and `MarkUp` loop over the backends by index and change the one
whose `addr` matches. In `Next`, repeat task 1's "take this one and move the
turn on" up to `len(rr.backends)` times, returning as soon as the backend
you took is healthy.

### Closer

`for i := range rr.backends` gives you the index, so
`rr.backends[i].healthy = false` changes the real element. In `Next`,
`for range len(rr.backends) { ... }` runs once per backend, and
`if b.healthy { return b.addr, nil }` stops at the first healthy one.

### Answer

```go
// MarkDown (MarkUp is the same with true)
rr.mu.Lock()
defer rr.mu.Unlock()
for i := range rr.backends {
	if rr.backends[i].addr == addr {
		rr.backends[i].healthy = false
	}
}

// Next
rr.mu.Lock()
defer rr.mu.Unlock()
for range len(rr.backends) {
	b := rr.backends[rr.next]
	rr.next = (rr.next + 1) % len(rr.backends)
	if b.healthy {
		return b.addr, nil
	}
}
return "", nil // task 3 fixes this line
```

## Task 3: Fail fast when every backend is down

If the loop from task 2 ends without returning, it has looked at every
backend and none was healthy. The line after the loop is where that case
goes.

### Closer

With zero backends, `for range len(rr.backends)` runs zero times, so nothing
is indexed and nothing is divided by zero. Return the sentinel error after
the loop.

### Answer

```go
for range len(rr.backends) {
	// ... as in task 2
}
return "", ErrNoHealthy
```

After a full lap, `rr.next` has moved on `len(rr.backends)` times, so it is
back where it started.

## Task 4: Probe every backend

Three steps: lock, clone `rr.backends`, unlock. Probe each copy into a
`[]bool`, one slot per backend. Lock again, copy each result into
`rr.backends[i].healthy`, and count the `true`s.

### Closer

`backends := slices.Clone(rr.backends)` (add `"slices"` to the imports).
`results := make([]bool, len(backends))`, then
`results[i] = probe(b.addr)` in a loop. After `rr.mu.Lock()` and
`defer rr.mu.Unlock()`, loop over `results` with its index.

### Answer

```go
rr.mu.Lock()
backends := slices.Clone(rr.backends)
rr.mu.Unlock()

results := make([]bool, len(backends))
for i, b := range backends {
	results[i] = probe(b.addr)
}

rr.mu.Lock()
defer rr.mu.Unlock()
healthy := 0
for i, ok := range results {
	rr.backends[i].healthy = ok
	if ok {
		healthy++
	}
}
return healthy
```
