# Mutexes: atomics and slow locks

Two improvements to shared state in a deploy service: atomic counters, which
need no lock at all, and a cache that stops holding its lock during slow
network calls.

**New here:** `sync/atomic` (`atomic.Int64`), and unlocking before slow work.

## Learn

### Recap

A `sync.Mutex` protects shared data: `Lock`, `defer Unlock`, touch the data.
Every access, reads included, must hold the lock. Hold it briefly.

### Atomic counters

For a single number that many goroutines update, `sync/atomic` is simpler
than a mutex. An `atomic.Int64` is a number whose operations can't interleave:

```go
import "sync/atomic"

type Server struct {
	requests atomic.Int64 // zero value is 0, ready to use
}

func (s *Server) Handle()         { s.requests.Add(1) }
func (s *Server) Requests() int64 { return s.requests.Load() }
```

| Method | Does |
|---|---|
| `Add(n)` | adds `n`, returns the new value |
| `Load()` | reads the value, as an `int64` |
| `Store(v)` | sets the value |

> **Why is it written like this?** **Go requires** you to use these methods:
> `atomic.Int64` is a struct, so `++`, `+` and returning it as an `int64` won't
> compile. Changing the field's type is enough to make every old `++` an error
> the compiler points at. **Convention:** atomic fields go in a struct used by
> pointer (`*Server`), never copied.

Use atomics for independent counters and flags (`atomic.Bool`,
`atomic.Uint64`, …). Use a mutex when several values must change together, or
for maps and slices.

### Don't hold a lock during slow work

```go
c.mu.Lock()
defer c.mu.Unlock()
v := fetch(key) // a 300ms API call, while every other caller waits
```

`defer` keeps the lock until the function returns, slow call included. Every
other goroutine, even one asking for a different key, waits. Instead, lock
only around the map access, in two short sections:

```text
lock → check map → unlock → slow call (no lock) → lock → store → unlock
```

Without `defer`, you write `Unlock()` yourself and must make sure every path
calls it. Two goroutines missing the same key at once may both call the slow
function. For a cache that's fine: the last store wins.

## Your task

### 1. Make the stats lock-free

**Goal:** change the `total` and `failed` fields of `DeployStats` to
`atomic.Int64`, and update `Record`, `Total` and `Failed` to match.

`Record(ok)` adds 1 to `total`, and also 1 to `failed` when `ok` is `false`.
`Total()` and `Failed()` return the counts. Many goroutines call `Record` at
once. Add `"sync/atomic"` to the imports.

```go
var s DeployStats
s.Record(true)
s.Record(false)
s.Record(true)
s.Total()  // => 3
s.Failed() // => 1
```

**Think it through:**
- Change the field types first, then fix each line the compiler complains
  about.
- Which method in the table replaces `++`, and which replaces a plain read?

### 2. Stop holding the lock during lookups

**Goal:** change `AMICache.Get(region, lookup)` so the lock is held only while
reading or writing `c.ids`, never while `lookup` runs.

On a hit it returns the cached ID without calling `lookup`. On a miss it calls
`lookup(region)`, stores the result in `c.ids`, and returns it. Lookups for
different regions must be able to run at the same time.

```go
c := NewAMICache()
c.Get("eu-west-1", lookup) // calls lookup, => "ami-eu-west-1"
c.Get("eu-west-1", lookup) // cached: lookup is not called again
```

**Think it through:**
- Which line keeps the lock held until `Get` returns?
- Map the steps in the diagram above onto the lines of `Get`. Where does
  each lock and unlock go?

## Learn more

- [sync/atomic](https://pkg.go.dev/sync/atomic)
- [Go by Example: Atomic counters](https://gobyexample.com/atomic-counters)
- [The Go memory model](https://go.dev/ref/mem)
