## Task 1: swap the types

Change the fields to `total atomic.Int64` and `failed atomic.Int64`, and add
`"sync/atomic"` to the imports. Then `s.total++` becomes `s.total.Add(1)`,
and `return s.total` becomes `return s.total.Load()`.

## Task 2: two short critical sections

Don't use `defer` for the first unlock, because you want to release the lock
in the middle of the function. Lock, read the map, unlock. If it was a hit,
return. Otherwise call `lookup` with no lock held, then lock again to store
the result.

## Task 2: almost the answer

```go
c.mu.Lock()
id, ok := c.ids[region]
c.mu.Unlock()
if ok {
	return id
}

id = lookup(region)

c.mu.Lock()
defer c.mu.Unlock()
c.ids[region] = id
return id
```
