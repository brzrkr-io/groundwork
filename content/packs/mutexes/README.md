# Mutexes

When several goroutines share one piece of data, such as a map of request
counts, they must take turns. A *mutex* (short for "mutual exclusion") is the
lock that makes them take turns.

**New here:** `sync.Mutex`, `sync.RWMutex`, `maps.Clone`.

## Learn

### Why goroutines must take turns

`hits[path]++` looks like one step, but it is three: read, add one, write
back. Two goroutines doing it at once can lose an update:

```text
goroutine A: read 5 ─────────── add 1 ─ write 6
goroutine B:        read 5 ─ add 1 ─ write 6
result: 6, not 7
```

With maps, Go may even crash: `fatal error: concurrent map writes`. Two
goroutines touching the same data at once, at least one writing, is a *data
race*. The tests in this module run with Go's race detector, which fails the
test with "race detected". *The race detector* covers it in depth.

### sync.Mutex

`Lock` waits until nobody else holds the lock, then takes it. `Unlock`
releases it. The code between them, the *critical section*, runs in one
goroutine at a time:

```go
type Inventory struct {
	mu    sync.Mutex // protects items
	items map[string]int
}

func (inv *Inventory) Remove(name string) {
	inv.mu.Lock()
	defer inv.mu.Unlock() // released when Remove returns
	delete(inv.items, name)
}
```

> **Why is it written like this?**
> - **Use pointer receivers** (`*Inventory`): a value receiver copies the
>   mutex, and a copied mutex is a separate lock. `go vet` warns about it.
> - **Convention:** the mutex sits just above the data it protects, with a
>   `// protects …` comment. `defer …Unlock()` goes on the line right after
>   `Lock`, so no `return` can skip it.
> - **Free:** the zero value of a mutex is unlocked; no setup needed.

### sync.RWMutex: many readers, one writer

Many goroutines may *read* a map at once, as long as nobody writes. A
`sync.RWMutex` has two kinds of lock:

| Method pair | Who can hold it at the same time | Use it in |
|---|---|---|
| `Lock` / `Unlock` | one writer, no readers | methods that change the data |
| `RLock` / `RUnlock` | any number of readers, no writer | methods that only read |

Reads (`RLock`) still wait while a writer holds `Lock`. That waiting is what
makes them safe.

### Keep critical sections small

Hold the lock only while touching the shared data, never during something
slow like a network call. Everyone else waits meanwhile.

### Don't hand out the protected map

A map value is a reference: returning `c.hits` gives the caller the **same**
map, which they can then use with no lock. Return a **copy**.
`maps.Clone(m)` from the `maps` package makes one:

```go
import "maps"

backup := maps.Clone(settings) // new map, same keys and values
```

## Your task

Your edge proxy counts requests per path. Many request goroutines call `Inc`,
and a metrics endpoint reads the counts. `HitCounter` keeps the counts in the
map `c.hits`, protected by the `sync.RWMutex` field `c.mu`.

### 1. Count a hit

**Goal:** make `Inc(path)` add 1 to `c.hits[path]` while holding the **write**
lock.

Many goroutines call it at once; no hit may be lost.

```go
c := NewHitCounter()
c.Inc("/login")
c.Inc("/login")
c.Inc("/health")
// c.hits is now map[string]int{"/login": 2, "/health": 1}
```

**Think it through:**
- Which method pair from the table is for changing data?
- The lock lives in `c`. How did the `Inventory` example reach its lock?

### 2. Read a count

**Goal:** make `Count(path)` return `c.hits[path]` while holding the **read**
lock.

A path that has never been hit returns 0.

```go
c.Count("/login")  // => 2
c.Count("/health") // => 0
```

**Think it through:**
- Which method pair is for reading only?
- What does a map give you for a key it doesn't have?

### 3. Take a snapshot

**Goal:** make `Snapshot()` return a **copy** of `c.hits`, made while holding
the read lock.

Changing the returned map must not change the counter. If a writer holds the
lock, `Snapshot` must wait for it. Add `"maps"` to the import block.

```go
snap := c.Snapshot() // => map[string]int{"/login": 2, "/health": 1}
snap["/login"] = 100
c.Count("/login")    // => still 2
```

**Think it through:**
- Is this a read or a write of `c.hits`?
- Which function from Learn makes the copy?

## Learn more

- [A Tour of Go: sync.Mutex](https://go.dev/tour/concurrency/9)
- [Go by Example: Mutexes](https://gobyexample.com/mutexes)
- [sync.RWMutex](https://pkg.go.dev/sync#RWMutex)
- [maps.Clone](https://pkg.go.dev/maps#Clone)
