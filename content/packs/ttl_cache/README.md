# TTL caching: expire on read

A *cache* keeps the answer to a slow question so you don't have to ask again.
A *TTL* (time to live) says how long that answer can be trusted. You'll build
a small generic cache that forgets entries when their TTL runs out.

**New here:** TTL expiry, `time.Time.Before`. (`var zero V` is from *Generics*; deleting from a map while you `range` over it from *Rate limiting: fix the bucket*.)

## Learn

### Why cache?

Ops tools ask the same slow questions again and again: "which region is
`web-1` in?", "what IP is `db.internal`?". A dashboard that refreshes every
5 seconds for 200 hosts makes 40 API calls a second for answers that almost
never change. A cache answers from memory instead.

### TTL and staleness

Cached answers go *stale*: a host moves region, a DNS record changes. A TTL
limits how stale an answer can get. Each entry stores the moment it expires,
next to its value:

```go
deadline := start.Add(5 * time.Minute) // a time.Time 5 minutes after start
```

`t.Before(u)` is `true` only when `t` is **strictly** earlier than `u`:

```go
nine.Before(ten)  // => true
ten.Before(ten)   // => false (equal is not "before")
```

In this cache an entry is **expired when the current time is at or after its
expiry time**. With a 60s TTL and a `Set` at 09:00:00, a `Get` at
09:00:59.999 is a hit, and a `Get` at exactly 09:01:00 is a miss.

> **Why is it written like this?** `Before` is a method, so you call it on
> one time and pass the other. Storing the expiry time (instead of the time
> the entry was saved) is the author's choice: every check becomes a single
> comparison.

### A generic cache

As in *Generics*, type parameters let one type work for any keys and values:

```go
regions := cache.New[string, string](time.Minute, time.Now)
ports := cache.New[int, bool](time.Minute, time.Now)
```

`K comparable` means the key type must work with `==`, which map keys need.
`V any` allows any value type.

When a lookup misses, `Get` must still return *some* `V`. Inside generic code
you don't know what `V` is, so you declare a variable and let Go fill in its
zero value:

```go
var empty T // "" if T is string, 0 if int, false if bool
```

> **Why is it written like this?** Go requires every return value to be
> present, even on a miss. `var x T` is the standard way to get "the zero
> value of a type parameter". The name (`zero`, `empty`) is your choice.

### Expire on read, and clean up in the background

Expired entries still use memory until someone deletes them. This cache uses
both common approaches:

| Approach | How | Good at |
|---|---|---|
| Expire on read | `Get` sees the entry is expired, deletes it, reports a miss. | Never returning stale data. |
| Background cleanup | Something calls `Cleanup` every minute to delete everything expired. | Freeing memory for keys nobody reads again. |

Real code runs `Cleanup` from a `time.Ticker` loop; the tests call it
directly. Go allows `delete` on a map while you `range` over that same map,
so one loop can check and delete.

### Locks and the clock are done

Every method already locks `c.mu` and defers the unlock, as in *Mutexes*, so
you only write the logic. `c.now` is the cache's clock: real code passes
`time.Now`, the tests pass a fake clock they move by hand. Always call
`c.now()`, never `time.Now()`, or the tests can't control time.

## Your task

Your fleet dashboard looks up each host's region from a slow inventory API.
Finish the cache that sits in front of it.

### 1. Store and look up entries

**Goal:** make `Set` save a value with an expiry time, and make `Get` return it only while it is fresh.

`Set(k, v)` stores an `entry` in `c.items[k]`. The entry holds `v` and its
expiry time: the current time (`c.now()`) plus `c.ttl`. Setting a key again
replaces both the value and the expiry.

`Get(k)` looks `k` up in `c.items`:

- Missing: return the zero value and `false`.
- Expired (now is at or after the expiry): delete it from `c.items`, then
  return the zero value and `false`.
- Otherwise: return the value and `true`.

```go
c := cache.New[string, string](time.Minute, clock)
c.Set("web-1", "eu-west-1") // at 09:00:00
c.Get("web-1")              // at 09:00:30 => "eu-west-1", true
c.Get("web-1")              // at 09:01:00 => "", false (and the entry is gone)
c.Get("db-1")               // => "", false
```

**Think it through:**
- Look at the `entry` struct: which two fields do you fill in, and from which names?
- "Fresh" means now is *before* the expiry. How do you say "not fresh"?
- Missing and expired both return the same thing. Can one `if` cover both?

### 2. Delete entries and count them

**Goal:** make `Delete` remove a key and `Len` return how many entries are stored.

`Delete(k)` removes `k`. Deleting a missing key does nothing (and must not
panic). `Len()` returns how many entries `c.items` holds, **including**
expired ones nobody has removed yet.

```go
c.Set("web-1", "eu-west-1")
c.Set("db-1", "us-east-1")
c.Len()          // => 2
c.Delete("db-1")
c.Len()          // => 1
```

**Think it through:**
- Which built-in function removes a map key, and which one counts a map's entries?

### 3. Clean up expired entries

**Goal:** make `Cleanup` delete every expired entry and return how many it deleted.

Use the same "expired" rule as `Get`. Call `c.now()` **once** at the start
and compare every entry against that one time.

```go
// TTL 60s: web-1 set at 09:00:00, db-1 set at 09:00:50
c.Cleanup() // at 09:01:00 => 1 (web-1 removed)
c.Len()     // => 1
c.Cleanup() // at 09:01:10 => 0
```

**Think it through:**
- You need a loop over the map and a counter that starts at 0.
- Which entries get deleted? The same ones `Get` treats as expired.

## Learn more

- [time.Time.Before](https://pkg.go.dev/time#Time.Before)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [Cloudflare: What is DNS TTL?](https://www.cloudflare.com/learning/cdn/glossary/time-to-live-ttl/)
