## Task 1: Fix the stale read

Look at the comparison in `Get`. What does `After` return when the two times
are exactly equal?

### Closer

`now.After(expiresAt)` is false when the times are equal, so the entry is
still served at its expiry time. "Expired" means "not before":
`!c.now().Before(...)`.

### Answer

```go
if !ok || !c.now().Before(e.expiresAt) {
```

## Task 2: Load on a miss, but don't cache errors

Before loading anything, try the cache. After loading, only a successful
result should go into the cache.

### Closer

Start `GetOrLoad` with `if v, ok := c.Get(k); ok { ... }` and return early on
a hit. After `v, err := load(k)`, call `c.Set(k, v)` only when `err == nil`.

### Answer

```go
func (c *Cache[K, V]) GetOrLoad(k K, load func(K) (V, error)) (V, error) {
	if v, ok := c.Get(k); ok {
		return v, nil
	}
	cl, leader := c.startCall(k)
	// ... the !leader branch (task 3) ...
	v, err := load(k)
	if err == nil {
		c.Set(k, v)
	}
	c.finishCall(k, cl, v, err)
	return v, err
}
```

## Task 3: Collapse concurrent loads

`startCall` must notice a load that is already running for `k`, and callers
that aren't the leader must wait for it instead of loading. The `done`
channel is closed when the leader finishes.

### Closer

In `startCall`, add a comma-ok lookup on `c.inflight` before the code that
builds a new call; `defer c.mu.Unlock()` already unlocks on every return. In
the `!leader` branch, receive from `cl.done`, then read `cl.val` and `cl.err`.

### Answer

```go
// startCall, before creating a new call
if cl, ok := c.inflight[k]; ok {
	return cl, false
}

// GetOrLoad
if !leader {
	<-cl.done
	return cl.val, cl.err
}
```

Closing `done` wakes every waiter at once, and `finishCall` stores the result
before it closes the channel.
