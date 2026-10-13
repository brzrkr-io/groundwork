## Task 1: Count a hit

`Inc` changes the map, so it needs the write lock on `c.mu` while it does that. Take the lock, make the change, and make sure the lock is released when the method returns.

### Closer

Use `c.mu.Lock()`, then `defer c.mu.Unlock()` right after it. The change itself is `c.hits[path]++`.

### Answer

```go
c.mu.Lock()
defer c.mu.Unlock()
c.hits[path]++
```

## Task 2: Read a count

`Count` only reads, so it uses the read lock, which lets many readers in at once. Reading a missing key from a map already gives 0.

### Closer

Same shape as task 1, but with `c.mu.RLock()` and `defer c.mu.RUnlock()`. Then return `c.hits[path]`.

### Answer

```go
c.mu.RLock()
defer c.mu.RUnlock()
return c.hits[path]
```

## Task 3: Take a snapshot

Returning `c.hits` itself would let callers use the map with no lock held. Return a copy instead, made while you hold the read lock.

### Closer

`maps.Clone(c.hits)` makes the copy. Add `"maps"` to the import block next to `"sync"`.

### Answer

```go
c.mu.RLock()
defer c.mu.RUnlock()
return maps.Clone(c.hits)
```

If the test says "race detected", check that every method that touches `c.hits` takes the lock first.
