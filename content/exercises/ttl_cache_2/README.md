# TTL caching: collapse concurrent loads

A cache in front of DNS lookups, with a stale-read bug to fix and a
`GetOrLoad` method that makes sure a popular key is looked up only once, even
when hundreds of requests want it at the same moment.

## Learn

### Recap

As in *TTL caching: expire on read*, each entry stores its value and an
`expiresAt` time, set to `now + ttl` by `Set`. An entry is **expired once
now >= expiresAt**, so `c.now().Before(e.expiresAt)` means "still fresh". The
clock is injected as `now func() time.Time`, and every method locks `c.mu`.

### Load on a miss

Callers usually don't want "get, and if it's missing, look it up, then set".
`GetOrLoad` does all three. You pass it a *loader*, a function that fetches
the real value:

```go
ip, err := c.GetOrLoad("db.internal", func(host string) (string, error) {
	return lookupIP(host) // the slow DNS call
})
```

Only successful results go in the cache. If the lookup fails and you cache
the failure, every caller gets the error for the whole TTL, even after DNS
recovers a second later.

### The cache stampede

`db.internal` is used by every request. Its entry expires. In the next
millisecond, 500 requests miss the cache at once, and all 500 call DNS. The
cache was protecting DNS from exactly this load, and it vanished at the worst
moment. This is a *cache stampede*.

The fix is *request collapsing* (Go's `golang.org/x/sync/singleflight`
package does this): the first caller to miss becomes the *leader* and runs
the loader. Callers that arrive while the load is running don't start their
own. They wait for the leader and share its result.

```text
caller 1 ── miss ── startCall: leader ── load(k) ──────── finishCall ── 10.0.0.7
caller 2 ── miss ── startCall: wait ─── <-cl.done ...... woken ──────── 10.0.0.7
caller 3 ── miss ── startCall: wait ─── <-cl.done ...... woken ──────── 10.0.0.7
```

A running load is a `call`. Waiters block on its `done` channel. Closing a
channel wakes **every** goroutine receiving from it, so after the leader
stores `val` and `err` and closes `done`, all the waiters can read the result.
`c.inflight` maps each key to its running call, and `finishCall` (already
written) does the storing, the closing and the cleanup. If the leader never
calls `finishCall`, the waiters wait forever.

### Testing it with synctest.Wait

You've seen `testing/synctest` give tests a fake clock. The collapsing test
also uses `synctest.Wait()`. It blocks until every other goroutine in the
test's bubble is *durably blocked*: stuck on a channel operation that only
another goroutine in the bubble can finish. The test starts 10 callers with a
loader that blocks on a `release` channel, calls `synctest.Wait()`, and then
checks the loader ran exactly once. A caller that is briefly waiting for
`c.mu` isn't durably blocked, so `Wait` keeps waiting until it gets the lock
and reaches `<-cl.done` or the loader. Then the test closes `release` and
checks every caller got the same answer.

The tests count loader calls with an `atomic.Int32` from `sync/atomic`:
`calls.Add(1)` and `calls.Load()` are safe from many goroutines.

## Your task

Your service discovery client resolves internal hostnames through a cache.
Fix it and stop it from stampeding DNS.

### 1. Fix the stale read

`Get` still returns an entry at the exact moment it expires. With a 30s TTL
and a `Set` at 09:00:00:

```text
Get at 09:00:29.999999999  => "10.0.0.7", true
Get at 09:00:30            => "", false   (currently returns the stale value)
```

Fix the comparison in `Get`.

### 2. Load on a miss, but don't cache errors

`GetOrLoad(k, load)` always calls `load` right now, and never caches anything.
(Ignore `startCall`, `leader` and `finishCall` for now: they are for task 3.)
Fix it:

- If `Get(k)` finds a fresh value, return it and a `nil` error. Don't call
  `load`, and don't call `startCall`.
- Otherwise, after `load(k)` returns a `nil` error, `Set` the value.
- Return exactly what `load` returned (the loaders in the tests return `""`
  with their errors).

```go
c.GetOrLoad("db.internal", resolve)   // calls resolve => "10.0.0.7", nil
c.GetOrLoad("db.internal", resolve)   // cached        => "10.0.0.7", nil
c.GetOrLoad("cache.internal", broken) // => "", errDNS, and nothing is cached
```

### 3. Collapse concurrent loads

`GetOrLoad` already registers its load with `startCall(k)` and publishes the
result with `c.finishCall(k, cl, v, err)`. But `startCall` always creates a
new call, so every caller becomes the leader and loads.

First fix `startCall(k)`. If `c.inflight[k]` already holds a call, return it
and `false`. Otherwise keep the existing code: create a call, store it in
`c.inflight[k]`, and return it and `true`.

Then, in `GetOrLoad`, fix the `!leader` branch. Someone else is loading, so
don't call `load`: wait with `<-cl.done`, then return `cl.val, cl.err`.

Ten concurrent `GetOrLoad("db.internal", ...)` calls must run the loader
**once**, and all ten get its value, or all ten get its error. Different keys
still load separately (in parallel).

## Learn more

- [testing/synctest](https://pkg.go.dev/testing/synctest)
- [golang.org/x/sync/singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)
- [Go spec: Close](https://go.dev/ref/spec#Close)
