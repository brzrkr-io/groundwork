# Mutexes

When several goroutines share one piece of data, such as a map of request
counts, they must take turns. A *mutex* (short for "mutual exclusion") is the
lock that makes them take turns.

## Learn

### Why goroutines must take turns

`hits[path]++` looks like one step, but it is three: read the old value, add
one, write it back. If two goroutines do this at the same moment, one update
can be lost:

```text
goroutine A: read 5 ─────────── add 1 ─ write 6
goroutine B:        read 5 ─ add 1 ─ write 6
result: 6, not 7
```

Maps are worse: Go may crash the whole program with `fatal error: concurrent
map writes`. Two goroutines touching the same data at the same time, with at
least one of them writing, is a *data race*. Gopherlings runs every test in
this module with Go's race detector, which spots data races and fails the
test. The next-but-one exercise, *The race detector*, is all about it.

### sync.Mutex

A `sync.Mutex` has two methods. `Lock` waits until nobody else holds the lock,
then takes it. `Unlock` releases it. Code between them runs in one goroutine
at a time. That code is the *critical section*:

```go
type HitCounter struct {
	mu   sync.Mutex
	hits map[string]int
}

func (c *HitCounter) Inc(path string) {
	c.mu.Lock()
	defer c.mu.Unlock() // released when Inc returns, whatever happens
	c.hits[path]++
}
```

- Put the mutex in the struct, next to the data it protects.
- Use pointer receivers (`*HitCounter`). A copied mutex is a different lock,
  and `go vet` warns about it.
- The zero value of a mutex is an unlocked mutex. No setup needed.
- `defer c.mu.Unlock()` right after `Lock` means you can't forget to unlock,
  even when the function has several `return`s.

### sync.RWMutex: many readers, one writer

Reading a map from many goroutines at once is fine, as long as nobody is
writing. A `sync.RWMutex` allows that:

| Method | Who can hold it at the same time |
|---|---|
| `Lock` / `Unlock` | one writer, and no readers |
| `RLock` / `RUnlock` | any number of readers, and no writer |

Use `Lock` in methods that change the data and `RLock` in methods that only
read it. For a dashboard that reads counters far more often than it writes
them, this means readers don't queue up behind each other.

### Keep critical sections small

Hold the lock only while you touch the shared data. Never hold it while doing
something slow, like a network call. Every other goroutine waits in the
meantime.

### Don't hand out the protected map

If a method returns `c.hits` itself, the caller can read or write it with no
lock held, and you're back to a data race. Return a **copy** instead.
`maps.Clone(m)` from the `maps` package makes one:

```go
import "maps"

copied := maps.Clone(original) // a new map with the same keys and values
```

## Your task

Your edge proxy counts requests per path. Many request goroutines call `Inc`,
and a metrics endpoint reads the counts. `HitCounter` uses a `sync.RWMutex`
named `mu`.

### 1. Count a hit

`Inc(path)` should add 1 to the count for `path`, holding the **write** lock
(`Lock`/`Unlock`).

```go
c := NewHitCounter()
c.Inc("/login")
c.Inc("/login")
// c now holds "/login": 2
```

### 2. Read a count

`Count(path)` should return the count for `path`, holding the **read** lock
(`RLock`/`RUnlock`). A path that has never been hit has a count of 0.

```go
c.Count("/login")  // => 2
c.Count("/health") // => 0
```

### 3. Take a snapshot

`Snapshot()` should return a **copy** of all the counts, holding the read
lock. Changing the returned map must not change the counter. Add `"maps"` to
the imports.

```go
snap := c.Snapshot() // => map[string]int{"/login": 2}
snap["/login"] = 100
c.Count("/login")    // => still 2
```

## Learn more

- [A Tour of Go: sync.Mutex](https://go.dev/tour/concurrency/9)
- [Go by Example: Mutexes](https://gobyexample.com/mutexes)
- [sync.RWMutex](https://pkg.go.dev/sync#RWMutex)
- [maps.Clone](https://pkg.go.dev/maps#Clone)
