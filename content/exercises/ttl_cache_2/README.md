# TTL caching: collapse concurrent loads

A cache in front of DNS lookups, with a stale-read bug to fix and a
`GetOrLoad` method that makes sure a popular key is looked up only once, even
when hundreds of requests want it at the same moment.

**New here:** loader functions, cache stampedes, request collapsing with a closed channel, `synctest.Wait`.

## Learn

### Recap

As in *TTL caching: expire on read*, each entry stores its value and an
`expiresAt` time (`now + ttl`). An entry is **expired once now is at or after
`expiresAt`**. The clock is `c.now`, and every method locks `c.mu`.

`t.After(u)` is the mirror of `t.Before(u)`: `true` only when `t` is
**strictly** later. Neither one is true when the two times are equal.

### Load on a miss

Callers usually don't want "get; if missing, look it up; then set".
`GetOrLoad` does all three. You pass it a *loader*: a function that fetches
the real value.

```go
port, err := ports.GetOrLoad("ssh", func(name string) (int, error) {
	return lookupPort(name) // the slow call
})
```

Only **successful** results go in the cache. If you cache a failure, every
caller gets the error for the whole TTL, even after DNS recovers a second
later.

> **Why is it written like this?** The `func(K) (V, error)` parameter type is
> required by `GetOrLoad`'s signature. Returning `(value, error)` with the
> error last is a Go convention. Writing the loader inline (instead of naming
> it) is the caller's choice.

### The cache stampede

`db.internal` is used by every request. Its entry expires. In the next
millisecond 500 requests miss at once, and all 500 call DNS. The cache
existed to protect DNS from exactly this. That's a *cache stampede*.

The fix is *request collapsing* (what `golang.org/x/sync/singleflight` does):
the first caller to miss becomes the *leader* and runs the loader. Callers
that arrive while it runs don't load. They wait and share the leader's result.

```text
caller 1 ── miss ── startCall: leader ── load(k) ──────── finishCall ── 10.0.0.7
caller 2 ── miss ── startCall: wait ─── blocked on done .. woken ─────── 10.0.0.7
caller 3 ── miss ── startCall: wait ─── blocked on done .. woken ─────── 10.0.0.7
```

### How waiting works

A running load is a `call`. It has a `done` channel plus `val` and `err`
fields. `c.inflight` maps each key to its running call.

Receiving from a channel blocks until something is sent **or the channel is
closed**. Closing wakes **every** receiver at once:

```go
<-ready // blocks here until someone runs close(ready)
```

`finishCall` (already written) stores `val` and `err`, removes the call from
`c.inflight`, and closes `done`. So after a waiter wakes up, the result is
already sitting in the call's fields.

> **Why is it written like this?** `<-ch` on its own line (ignoring the
> value) is plain Go: a receive is an expression you may discard. Using
> `close` as a broadcast is a Go convention. `chan struct{}` is convention
> for "signal only, no data".

### Reading the tests

`synctest.Wait()` blocks until every other goroutine in the test is *durably
blocked*: stuck on a channel that only another goroutine in the test can
release. The collapsing test starts 10 callers whose loader blocks on a
`release` channel, calls `Wait`, checks the loader ran **once**, then closes
`release` and checks all 10 got the same answer.

`atomic.Int32` counts loader calls safely: `calls.Add(1)`, `calls.Load()`.

## Your task

Your service discovery client resolves internal hostnames through a cache.
Fix it and stop it from stampeding DNS.

### 1. Fix the stale read

**Goal:** make `Get` report a miss at the exact moment an entry expires.

With a 30s TTL and a `Set` at 09:00:00:

```text
Get at 09:00:29.999999999  => "10.0.0.7", true
Get at 09:00:30            => "", false   (currently returns the stale value)
```

**Think it through:**
- Read the `if` in `Get`. At exactly 09:00:30, is now *after* `expiresAt`?
- You want "expired" to include "equal". Which method is `true` for "still
  fresh", and how do you flip it?

### 2. Load on a miss, but don't cache errors

**Goal:** make `GetOrLoad` return a fresh cached value without loading, and cache only successful loads.

Ignore `startCall`, `leader` and `finishCall` for now: they're for task 3.

- If `Get(k)` finds a fresh value, return it and a `nil` error. Don't call
  `load`, and don't call `startCall`.
- Otherwise, run `load(k)`. If its error is `nil`, `Set` the value.
- Return exactly what `load` returned (the test loaders return `""` with
  their errors).

```go
c.GetOrLoad("db.internal", resolve)   // calls resolve => "10.0.0.7", nil
c.GetOrLoad("db.internal", resolve)   // cached        => "10.0.0.7", nil
c.GetOrLoad("cache.internal", broken) // => "", errDNS, and nothing is cached
```

**Think it through:**
- The cache check must come first, above the line that calls `startCall`.
- After `load` returns in the leader path, where do you check `err` before saving?

### 3. Collapse concurrent loads

**Goal:** make concurrent `GetOrLoad` calls for the same key run the loader only once and share its result.

Two places change:

- **`startCall(k)`**: if `c.inflight[k]` already holds a call, return that
  call and `false`. Otherwise keep the existing code (create, store, return
  it and `true`).
- **The `!leader` branch in `GetOrLoad`**: someone else is loading, so don't
  call `load`. Wait until that call is finished, then return the value and
  error stored on it.

Ten concurrent `GetOrLoad("db.internal", ...)` calls run the loader **once**.
All ten get its value, or all ten get its error. Different keys still load
separately.

**Think it through:**
- In `startCall`, use the two-value map lookup (`value, ok := m[key]`).
- In the waiter branch: which field of `cl` do you receive from to wait, and
  which two fields hold the result?

## Learn more

- [testing/synctest](https://pkg.go.dev/testing/synctest)
- [golang.org/x/sync/singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)
- [Go spec: Close](https://go.dev/ref/spec#Close)
