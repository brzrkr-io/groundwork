# TTL caching: expire on read

A *cache* keeps the answer to a slow question so you don't have to ask again.
A *TTL* (time to live) says how long that answer can be trusted. You'll build
a small generic cache that forgets entries when their TTL runs out.

## Learn

### Why cache?

Ops tools ask the same slow questions over and over: "which region is
`web-1` in?" (an inventory API call), "what IP is `db.internal`?" (a DNS
lookup), "what tags does this instance have?" (a cloud API call that is
rate-limited). If a dashboard refreshes every 5 seconds for 200 hosts, that's
40 API calls a second for answers that almost never change. A cache answers
from memory instead.

### TTL and staleness

Cached answers go *stale*: a host moves region, a DNS record changes. A TTL
limits how stale an answer can get. Each entry remembers when it expires:

```go
expiresAt := c.now().Add(c.ttl) // stored with the value in Set
```

In this cache an entry is **expired when the current time is at or after
`expiresAt`**. With a 60s TTL and a `Set` at 09:00:00, a `Get` at 09:00:59.999
is a hit, and a `Get` at exactly 09:01:00 is a miss. `t.Before(u)` reports
whether `t` is strictly earlier than `u`, so "still fresh" is:

```go
c.now().Before(e.expiresAt) // true: fresh; false: expired
```

Short TTLs mean fresher data but more slow calls. Long TTLs mean fewer calls
but staler data. DNS records carry their own TTLs for exactly this reason.

### A generic cache

As in *Generics*, type parameters let one type work for any keys and values:

```go
type Cache[K comparable, V any] struct { ... }

regions := cache.New[string, string](time.Minute, time.Now)
ports := cache.New[int, bool](time.Minute, time.Now)
```

`K comparable` means the key type must work with `==`, which map keys need.
`V any` allows any value type. When a lookup misses, `Get` must still return
*some* `V`. `var zero V` declares a variable holding the zero value of
whatever `V` is: `""` for strings, `0` for ints, `false` for bools.

### Expire on read, or clean up in the background?

Expired entries still take up memory until someone deletes them. There are
two common approaches, and this cache uses both:

| Approach | How | Good at |
|---|---|---|
| Expire on read | `Get` notices the entry is expired, deletes it and reports a miss. | Never returning stale data. |
| Background cleanup | A goroutine calls `Cleanup` every minute or so to delete everything expired. | Freeing memory for keys nobody reads again. |

Real code would run `Cleanup` from a `time.Ticker` loop. The tests call it
directly. Deleting map entries while you `range` over the same map is allowed
in Go, so `Cleanup` can loop and delete in one pass.

### Safe for many goroutines

HTTP handlers run in many goroutines at once, and they all share one cache.
Every method already locks `c.mu` and defers the unlock, as in *Mutexes*, so
you only write the logic. The tests run with the race detector.

### An injected clock

`now func() time.Time` is the cache's clock. Real code passes `time.Now`, and
the tests pass a fake clock that they move forward by hand, so expiry is
tested without waiting.

## Your task

Your fleet dashboard looks up each host's region from a slow inventory API.
Finish the cache that sits in front of it.

### 1. Store and look up entries

In `Set`, store an `entry` holding `v` and `expiresAt` (`c.now()` plus
`c.ttl`) in `c.items[k]`. Setting a key again replaces both the value and the
expiry.

In `Get`, look `k` up in `c.items`:

- Missing: return the zero value and `false`.
- Expired (now is at or after `expiresAt`): delete it from `c.items`, then
  return the zero value and `false`.
- Otherwise return the value and `true`.

```go
c := cache.New[string, string](time.Minute, clock)
c.Set("web-1", "eu-west-1") // at 09:00:00
c.Get("web-1")              // at 09:00:30 => "eu-west-1", true
c.Get("web-1")              // at 09:01:00 => "", false (and the entry is gone)
c.Get("db-1")               // => "", false
```

### 2. Delete entries and count them

`Delete(k)` removes `k`. Deleting a missing key does nothing. `Len()` returns
how many entries `c.items` holds, **including** expired ones nobody has
removed yet.

```go
c.Set("web-1", "eu-west-1")
c.Set("db-1", "us-east-1")
c.Len()          // => 2
c.Delete("db-1")
c.Len()          // => 1
```

### 3. Clean up expired entries

`Cleanup()` deletes every expired entry (same rule as `Get`) and returns how
many it deleted. Call `c.now()` once at the start and compare every entry
against that time.

```go
// TTL 60s: web-1 set at 09:00:00, db-1 set at 09:00:50
c.Cleanup() // at 09:01:00 => 1 (web-1 removed)
c.Len()     // => 1
c.Cleanup() // at 09:01:10 => 0
```

## Learn more

- [time.Time.Before](https://pkg.go.dev/time#Time.Before)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [Cloudflare: What is DNS TTL?](https://www.cloudflare.com/learning/cdn/glossary/time-to-live-ttl/)
