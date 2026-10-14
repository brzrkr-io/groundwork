# The race detector: hidden races

The races in *The race detector* sat right on the `+=` or `append` line.
Real ones hide better: a read just above a `Wait`, an `err` shared by every
goroutine, a "load it the first time" cache. Hunt down three of them.

**New here:** `sync.Once`; races that hide.

## Learn

### Recap

A data race is two goroutines using the same variable at the same time, with
at least one writing, and nothing making them take turns. The
`WARNING: DATA RACE` report's two stacks point at the racing lines. Fix it
with an atomic, a mutex, or by not sharing.

### Where races hide

| Pattern | Why it races |
|---|---|
| Reading results **before** `wg.Wait()` | Goroutines may still be writing. `Wait` is what makes their writes safe to read. |
| `err = e` inside goroutines | Every goroutine writes the same `err` variable. |
| `if c.token == "" { c.token = fetch() }` | Two goroutines both see `""`, and both write. |

### Don't share: one slot each

Give each goroutine its own index in a slice made before the loop. No two
goroutines touch the same element, so there is nothing to lock:

```go
pings := make([]time.Duration, len(hosts))
for i, host := range hosts {
	wg.Go(func() {
		pings[i] = ping(host) // goroutine i only writes slot i
	})
}
wg.Wait()
// now read pings
```

> **Why is it written like this?** Go requires the slice to have its full
> length (`make([]T, n)`) before goroutines write `pings[i]`; `append` would
> share the slice again. Reading only after `Wait` is required, not style.

### sync.Once

`sync.Once` runs a function exactly once, however many goroutines ask at the
same time. Callers that arrive while it runs wait until it has finished:

```go
type Config struct {
	once   sync.Once
	values map[string]string
}

func (c *Config) Get(key string) string {
	c.once.Do(func() {
		c.values = loadConfig() // runs once, on the first call
	})
	return c.values[key] // safe: every caller waits for Do to finish
}
```

> **Why is it written like this?** `Do` takes a function with no parameters
> and no results, so the work goes in an anonymous `func() { ... }` that sets
> a field. A `sync.Once` must not be copied, so Go code uses pointer
> receivers (`*Config`). Field names are the author's choice.

## Your task

A container release tool has three data races. Fix each one so its test
passes with no race report, and keep the work concurrent.

### 1. Read after waiting

**Goal:** make `DiskUsage(dirs, size)` return each size (in `dirs` order)
and the total, without reading too early.

It measures every directory at the same time, but adds up the total before
the goroutines are done.

```go
DiskUsage([]string{"/var/log", "/var/lib/docker"}, size)
// => []int{120, 4000}, 4120
```

**Think it through:**
- The goroutines are fine. Which line comes in the wrong place?

### 2. Stop sharing err

**Goal:** make `PushAll(images, push)` return `nil` if every push succeeded,
otherwise the error of the **first image in `images` order** whose push
failed.

Every goroutine writes the same `err` now. Give each goroutine its own slot,
and pick the first non-nil error after waiting.

```go
// push fails for "api:1.4" and "web:2.0"
PushAll([]string{"db:9.1", "api:1.4", "web:2.0"}, push)
// => the error from "api:1.4" (it comes before "web:2.0" in the list)
```

**Think it through:**
- What type holds one error per image? How long must it be?
- You need each image's position. What does the `range` give you besides
  the value?

### 3. Fetch the token once

**Goal:** make `TokenCache.Token()` call `c.fetch()` exactly once, however
many goroutines call it at the same time, and return its result every time.

The struct already has a `sync.Once` field, `c.once`. Use it.

```go
c := NewTokenCache(fetch)
c.Token() // calls fetch, => "tok-123"
c.Token() // => "tok-123", fetch is not called again
```

**Think it through:**
- Which lines does the `if` do that `c.once.Do` can do instead?
- After `Do` returns, what do you return?

## Learn more

- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [sync.Once](https://pkg.go.dev/sync#Once)
- [The Go memory model](https://go.dev/ref/mem)
