# The race detector: hidden races

The races in *The race detector* sat right on the `+=` or `append` line.
Real ones hide better: a read just above a `Wait`, an `err` variable shared
by every goroutine, a "load it the first time" cache. Hunt down three of them.

## Learn

### Recap

A data race is two goroutines using the same variable at the same time, with
at least one writing, and nothing making them take turns. The test output
shows a `WARNING: DATA RACE` report: the two stacks point at the racing
lines in your file. Fix races with an atomic, a mutex, or by not sharing
(each goroutine writes its own slot, and you combine the slots after
`wg.Wait()`).

### Where races hide

| Pattern | Why it races |
|---|---|
| Reading results **before** `wg.Wait()` | The goroutines may still be writing. `Wait` is what makes their writes visible. |
| `err = e` inside goroutines | Every goroutine writes the same `err` variable. |
| `if c.token == "" { c.token = fetch() }` | Two goroutines both see `""`, and both write. |

### sync.Once

`sync.Once` runs a function exactly once, however many goroutines ask at the
same time. Callers that arrive while it is running wait until it has
finished:

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

Like a mutex, a `sync.Once` must not be copied, so use pointer receivers.

## Your task

A container release tool has three data races. Fix each one so its test
passes with no race report, and keep the work concurrent.

### 1. Read after waiting

`DiskUsage(dirs, size)` measures every directory at the same time and
returns each size (in the order of `dirs`) plus the total. It adds up the
total too early. Fix it.

```go
DiskUsage([]string{"/var/log", "/var/lib/docker"}, size)
// => []int{120, 4000}, 4120
```

### 2. Stop sharing err

`PushAll(images, push)` pushes every image at the same time. It returns
`nil` if every push succeeded. Otherwise it returns the error of the **first
image in `images` order** whose push failed. Right now every goroutine writes
the same `err` variable. Give each goroutine its own slot instead, and pick
the first non-nil error after waiting.

```go
// push fails for "api:1.4" and "web:2.0"
PushAll([]string{"db:9.1", "api:1.4", "web:2.0"}, push)
// => the error from "api:1.4" (it comes before "web:2.0" in the list)
```

### 3. Fetch the token once

`TokenCache.Token()` returns a registry token, fetching it with `c.fetch()`
the first time it is called. Many goroutines call `Token` at once. They race
on `c.token`, and `fetch` may run several times. Use the `sync.Once` field
`c.once` so `fetch` runs exactly once and every caller gets its result.

```go
c := NewTokenCache(fetch)
c.Token() // calls fetch, => "tok-123"
c.Token() // => "tok-123", fetch is not called again
```

## Learn more

- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [sync.Once](https://pkg.go.dev/sync#Once)
- [The Go memory model](https://go.dev/ref/mem)
