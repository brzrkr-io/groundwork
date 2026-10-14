# DNS: TTLs and a migration check

You moved `api.example.com` to a new server and changed its DNS record, yet
some users still reach the old one. That's the TTL at work. Here you fix a
small TTL-aware cache, then write the check that says when a migration is
really finished.

**New here:** converting between seconds and `time.Duration`, passing a clock function in, reading TTLs with `dig`.

## Learn

### Recap: TTLs

Every DNS answer carries a *TTL* (time to live): how many seconds a resolver
may keep it. A resolver that fetched `api.example.com → 203.0.113.10, TTL 300`
at 12:00 keeps answering `203.0.113.10` until 12:05, whatever the record says
now. So the last old copy disappears one old TTL after a change, at the latest.

### Why this pack has its own `Record` type

`LookupHost`, from *DNS: lookups, timeouts and failure types*, returns only
addresses: Go's resolver throws the TTL away. Reading it would mean building
DNS messages yourself (with `golang.org/x/net/dns/dnsmessage`). So this pack
defines a small `Record` struct, with `TTL` in seconds as DNS sends it.

### Seconds and `time.Duration`

A `time.Duration` counts **nanoseconds**, so converting a plain number gives
you that many nanoseconds:

| Code | Value |
|---|---|
| `time.Duration(90)` | `90ns` |
| `time.Duration(90) * time.Second` | `1m30s` |
| `int(150 * time.Second / time.Second)` | `150` |
| `int(2500 * time.Millisecond / time.Second)` | `2` (rounds down) |

> **Why is it written like this?** Go requires both sides of `*` to have the
> same type, so the `int` must become a `time.Duration` first. Dividing one
> `Duration` by another gives a count, still typed `Duration`; `int(...)`
> turns it into an ordinary number.

### Passing the clock in

Calling `time.Now()` inside the cache would make the tests wait real minutes.
So `NewCache` takes a `now func() time.Time`, as *Port checks: wait for the
database* took a `sleep` function. Real code passes `time.Now`; the tests pass
a closure over a variable they move forward, as in *Fakes: a fake clock for
retries*:

```go
fake := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
clock := func() time.Time { return fake }
fake = fake.Add(time.Minute) // clock() now returns 12:01
```

### Reading TTLs with `dig`

`dig +noall +answer` prints only the answer lines; the second column is the
TTL in seconds:

```text
$ dig +noall +answer api.example.com
api.example.com.	247	IN	A	203.0.113.10
```

Ask the same resolver again and the TTL counts down, just as your `Get` will.
During a migration, ask each resolver your users rely on with
`dig @10.0.0.2 +short api.example.com`. Step **1. DNS** of the *Networking
from the ground up* guide shows how to ask the domain's own servers.

## Your task

Your team keeps a tiny DNS cache in a deploy tool, and wants a migration
check it can run every minute after a DNS change.

### 1. Fix the expiry time

**Goal:** make `ExpiresAt(r, fetched)` return the moment `r.TTL` **seconds** after `fetched`.

Right now every copy expires a few nanoseconds after it was fetched, so the
cache never gets a hit.

```go
fetched := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
ExpiresAt(Record{"api.example.com", "203.0.113.10", 300}, fetched)
// => 2026-10-13 12:05:00 +0000 UTC
```

**Think it through:**
- Which unit is `r.TTL` in, and which unit does `time.Duration` count?
- Which row of the table in the Learn section fixes that?

### 2. Count down the TTL in `Get`

**Goal:** make `Get(name)` return the cached record with its `TTL` set to the seconds left, or `false` once it has expired.

`Put` and the "not cached" case are done. Finish the rest:

- If the clock (`c.now()`) is not before the entry's `expires`, the copy has expired: return `Record{}`, `false`.
- Otherwise return the record with `TTL` set to the whole seconds left until `expires`, rounded down, and `true`.

```go
c.Put(Record{"api.example.com", "203.0.113.10", 300}) // at 12:00:00
c.Get("api.example.com") // at 12:00:00   => {api.example.com 203.0.113.10 300}, true
c.Get("api.example.com") // at 12:00:01.5 => {api.example.com 203.0.113.10 298}, true
c.Get("api.example.com") // at 12:01:40   => {api.example.com 203.0.113.10 200}, true
c.Get("api.example.com") // at 12:05:00   => {}, false
```

**Think it through:**
- Which method compares two times, and which one subtracts them?
- `e.record` is a copy, so you can change its `TTL` before returning it.

### 3. Check a migration

**Goal:** make `CheckMigration(m, answers, now)` return one status line saying whether the migration is done, still waiting for caches, or stuck.

- `m.Old` is the record before the change, `m.NewAddr` the new address, `m.Changed` when you changed it.
- Each `Answer` is what one resolver returns right now.
- A resolver is *stale* when its `Addr` is `m.Old.Addr`. Keep stale resolvers in the order of `answers`, joined with `", "`.
- The last old copy expires at `ExpiresAt(m.Old, m.Changed)`. Format it with `time.RFC3339`.

Return exactly one of these, checked in this order:

| When | Line |
|---|---|
| no resolver is stale | `done: all N resolvers return NEWADDR` (N is how many answers there are) |
| `now` is before the expiry | `waiting: OLDADDR still cached on STALE until EXPIRY (LEFT left)` (LEFT is the `time.Duration` from `now` to the expiry) |
| otherwise | `stuck: OLDADDR still served by STALE after its TTL ran out at EXPIRY` |

```go
m := Migration{
	Old:     Record{"api.example.com", "203.0.113.10", 300},
	NewAddr: "198.51.100.7",
	Changed: time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC),
}
answers := []Answer{{"10.0.0.2", "203.0.113.10"}, {"8.8.8.8", "198.51.100.7"}, {"1.1.1.1", "203.0.113.10"}}

CheckMigration(m, answers, m.Changed.Add(time.Minute))
// => "waiting: 203.0.113.10 still cached on 10.0.0.2, 1.1.1.1 until 2026-10-13T12:05:00Z (4m0s left)"
CheckMigration(m, answers, m.Changed.Add(10*time.Minute))
// => "stuck: 203.0.113.10 still served by 10.0.0.2, 1.1.1.1 after its TTL ran out at 2026-10-13T12:05:00Z"
```

*Stuck* is worth a page: the record wasn't changed everywhere, or that
resolver keeps answers longer than their TTL.

**Think it through:**
- Collect the stale resolvers first; how many there are decides the first row.
- Which helper from task 1 gives the expiry, and what does `fmt.Sprintf` print for a `time.Duration` with `%v`?

## Learn more

- [time.Duration](https://pkg.go.dev/time#Duration)
- [Amazon Route 53: values for simple records (TTL)](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/resource-record-sets-values-basic.html)
- [golang.org/x/net/dns/dnsmessage](https://pkg.go.dev/golang.org/x/net/dns/dnsmessage)
- [Time to live (Wikipedia)](https://en.wikipedia.org/wiki/Time_to_live)
