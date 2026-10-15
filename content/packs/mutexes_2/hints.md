## Task 1: Make the stats lock-free

`s.total++` on a plain `int64` is a data race when many goroutines call it. The `sync/atomic` package has a number type whose operations are safe from many goroutines.

### Closer

Change the fields to `total atomic.Int64` and `failed atomic.Int64`, and add `"sync/atomic"` to the imports. Then `s.total++` becomes `s.total.Add(1)`, and reads use `.Load()`.

### Answer

```go
type DeployStats struct {
	total  atomic.Int64
	failed atomic.Int64
}

func (s *DeployStats) Record(ok bool) {
	s.total.Add(1)
	if !ok {
		s.failed.Add(1)
	}
}

func (s *DeployStats) Total() int64  { return s.total.Load() }
func (s *DeployStats) Failed() int64 { return s.failed.Load() }
```

## Task 2: Stop holding the lock during lookups

`defer c.mu.Unlock()` keeps the lock until `Get` returns, including during the slow `lookup`. You need two short locked sections: one to check the map, one to store the result, with `lookup` in between.

### Closer

Drop the `defer`. Lock, read `id, ok := c.ids[region]`, then `c.mu.Unlock()`. If `ok`, return. Otherwise call `lookup` with no lock held, then lock again to store.

### Answer

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

The lock only covers the map access, so lookups for different regions run at the same time.
