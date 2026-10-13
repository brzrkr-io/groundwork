# Mutexes: atomics and slow locks

Two improvements to shared state in a deploy service: atomic counters, which
need no lock at all, and a cache that stops holding its lock during slow
network calls.

## Learn

### Recap

A `sync.Mutex` (or `sync.RWMutex`) protects shared data: `Lock`, `defer
Unlock`, touch the data. Every access, reads included, must hold the lock.
Keep the time you hold it short.

### Atomic counters

For a single number that many goroutines update, the `sync/atomic` package is
simpler than a mutex. `atomic.Int64` is a number whose operations can't
interleave:

```go
import "sync/atomic"

type Stats struct {
	requests atomic.Int64 // the zero value is 0, ready to use
}

func (s *Stats) Record()         { s.requests.Add(1) }
func (s *Stats) Requests() int64 { return s.requests.Load() }
```

| Method | Does |
|---|---|
| `Add(n)` | adds `n` and returns the new value |
| `Load()` | reads the value |
| `Store(v)` | sets the value |

Every read and write must go through these methods. A plain `s.requests++`
on an `int64` is a data race. There are also `atomic.Int32`, `atomic.Uint64`
and `atomic.Bool`.

Use atomics for independent counters and flags. Use a mutex when several
values must change together, or for maps and slices.

### Don't hold a lock during slow work

```go
c.mu.Lock()
defer c.mu.Unlock()
id := lookup(region) // a 300ms API call, while every other caller waits
```

Every goroutine that wants the cache, even for a different region, now waits
for that API call. Hold the lock only around the map access, and do the slow
part without it:

```text
lock → check map → unlock → slow lookup (no lock held) → lock → store → unlock
```

A side effect: two goroutines that miss the same key at the same moment might
both call `lookup`. For a cache, that is usually fine. The last one to store
its answer wins.

## Your task

### 1. Make the stats lock-free

`DeployStats` counts deploys from many goroutines, but its fields are plain
`int64`s, so the race detector fails the test. Change both fields to
`atomic.Int64` and update the methods to use `Add` and `Load`. Add
`"sync/atomic"` to the imports.

```go
var s DeployStats
s.Record(true)
s.Record(false)
s.Record(true)
s.Total()  // => 3
s.Failed() // => 1
```

### 2. Stop holding the lock during lookups

`AMICache.Get(region, lookup)` returns the cached image ID for a region, and
calls `lookup(region)` (slow) to fill the cache on a miss. It holds the lock
during `lookup`, so lookups for different regions can't run at the same time.

Change `Get` so the lock is held only while reading or writing `c.ids`, never
while `lookup` runs. A region that is already cached must not call `lookup`
again.

```go
c := NewAMICache()
c.Get("eu-west-1", lookup) // calls lookup, => "ami-eu-west-1"
c.Get("eu-west-1", lookup) // cached: lookup is not called again
```

## Learn more

- [sync/atomic](https://pkg.go.dev/sync/atomic)
- [Go by Example: Atomic counters](https://gobyexample.com/atomic-counters)
- [The Go memory model](https://go.dev/ref/mem)
