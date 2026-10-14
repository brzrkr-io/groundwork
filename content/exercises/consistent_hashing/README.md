# Consistent hashing: a ring with virtual nodes

*Consistent hashing* spreads keys across servers so that adding or removing
one server moves only a small share of the keys. Cache clients, sharded
databases and load balancers (Envoy's "ring hash") all use it. You'll grow
the hash ring into the real thing: many points per server, and servers that
can leave.

**New here:** virtual nodes, `slices.DeleteFunc`, `maps.DeleteFunc`.

## Learn

### Recap: the ring

As in *Hash rings: find a key's node*, every node sits on a circle of
`uint32` numbers, `r.points` holds the points sorted from low to high, and
`r.owners` maps a point to its node. A key belongs to the node at the first
point **at or after** its hash, found with `slices.BinarySearch`; a key past
the last point wraps around to the first. Adding a node moves only the keys
between it and the point before it.

The hash here is the same `crc32.ChecksumIEEE`, wrapped in a small
`hash(s string) uint32` function in the starter.

### Virtual nodes

With one point per node, the gaps between points are uneven, so one node
can own most of the circle. On the ring from *Hash rings: find a key's
node*, `collector-b` got only 11 of the 100 hosts `web-00` to `web-99`.
Similar names make it worse with crc32. Here is how 1000 keys spread over
`cache-1`, `cache-2` and `cache-3` with the hash in this exercise:

| points per node | cache-1 | cache-2 | cache-3 |
|---|---|---|---|
| 1 | 4 | 4 | 992 |
| 10 | 786 | 166 | 48 |
| 100 | 372 | 239 | 389 |

The fix is to put each node on the ring many times. Each point is a
*virtual node* (also called a *replica*). Each one is the hash of a label
made from the node name and a number: `db-a#0`, `db-a#1`, `db-a#2`, ... Many
small arcs average out.

```text
one point each:      db-a ·························· db-b ··· db-c ·····
three points each:   db-a ··· db-c ··· db-b ··· db-a ··· db-c ··· db-b ···
```

**Why the separator matters.** Without the `#`, node `cache-1` replica 12
is `"cache-1" + "12"` = `cache-112`, and node `cache-11` replica 2 is
`"cache-11" + "2"` = `cache-112` too. Same label, same point: two nodes
fight over one spot. Hostnames never contain `#`, so with it every label is
unique.

> **Why is it written like this?** Go never turns an `int` into its digits
> by itself, so you need `strconv.Itoa(i)`, as in *Type conversion*. The `#`
> is the author's choice: any character that never appears in a node name
> works.

### Deleting by a condition

When a node leaves, every one of its points must go. `slices.DeleteFunc`
removes every element for which your function returns `true`, keeping the
rest in order:

```go
hosts := []string{"old-web-1", "web-2", "old-db-1", "web-3"}
hosts = slices.DeleteFunc(hosts, func(h string) bool {
	return strings.HasPrefix(h, "old-")
})
// hosts is now ["web-2" "web-3"]
```

`maps.DeleteFunc` does the same for map entries. Your function gets each key
and its value:

```go
ttls := map[string]int{"web-2": 30, "db-1": 0, "web-3": 0}
maps.DeleteFunc(ttls, func(host string, secs int) bool {
	return secs == 0
})
// ttls is now map[web-2:30]
```

> **Why is it written like this?** `slices.DeleteFunc` returns the shorter
> slice, and Go requires you to store it (`hosts = ...`), or `hosts` keeps
> its old length. A map is changed in place, so `maps.DeleteFunc` returns
> nothing. The `func(...) bool { ... }` is a function literal, as in
> *Functions as values*.

### No mutex this time

This `Ring` is not safe for concurrent use, to keep the focus on the ring
itself. That's the author's choice: a real one adds a `sync.RWMutex`, as in
*Mutexes*, with `Get` taking the read lock.

## Your task

Your CI runners keep build artifacts on three cache servers, `cache-1` to
`cache-3`. With `hash % 3`, adding `cache-4` sent three quarters of the
lookups to the wrong server, and every build slowed down. Build a hash ring
with virtual nodes, so only about a quarter of the keys move and every
server gets a fair share.

### 1. Place nodes on the ring

**Goal:** make `Add` put each node on the ring at `r.replicas` points.

For every node, and every `i` from `0` to `r.replicas - 1`:

- The label is the node name, then the fixed text `#`, then `i`:
  `cache-1#0`, `cache-1#1`, ...
- The point is `hash(label)`.
- If `r.owners` already has that point, skip it. That way, adding a node
  twice changes nothing.
- Otherwise record the node in `r.owners` and append the point to
  `r.points`.

When `Add` returns, `r.points` must be sorted from low to high, also after
the second, third, ... call. Add `"slices"` and `"strconv"` to the imports.

```go
r := New(2)
r.Add("cache-1", "cache-2")
r.points            // => [134980999 172825566 2102012744 2131531025]
r.owners[134980999] // => "cache-1" (the hash of "cache-1#1")
r.Add("cache-1")    // already on the ring: nothing changes
```

**Think it through:**
- In *Hash rings: find a key's node*, `Add` had one loop. Which second loop
  does each node need now, and which goes inside which?
- Where in the label do the node and the number go, and what's between
  them?

### 2. Find a key's node

**Goal:** make `Get(key)` return the node at the first point at or after
`hash(key)`, wrapping around after the last point.

This is the same rule as in *Hash rings: find a key's node*; there's no
`Next` helper this time, so the binary search goes inside `Get`.

- If the ring has no points, return `""`.
- Find the first point that is **at or after** `hash(key)`. A key that
  hashes exactly onto a point belongs to that point.
- If there is no such point (the key hashes past the last point), wrap
  around and use the first point, `r.points[0]`.
- Return that point's owner from `r.owners`.

With one point per node, the ring from `New(1)` and
`Add("cache-1", "cache-2", "cache-3")` looks like this:

| point | owner |
|---|---|
| 2089304447 | cache-3 |
| 2102012744 | cache-2 |
| 2131531025 | cache-1 |

```go
r.Get("api/v1.4.2.tar.gz") // hash  573056911 => "cache-3"
r.Get("build-127.tar.gz")  // hash 2097901471 => "cache-2"
r.Get("cache-1#0")         // hash 2131531025 => "cache-1" (exactly on a point)
r.Get("web/v2.0.0.tar.gz") // hash 3969333627 => "cache-3" (past the last point: wraps)
New(100).Get("build-65.tar.gz") // => "" (empty ring)
```

The tests also check the whole idea with 100 points per node and the 1000
keys `build-0.tar.gz` to `build-999.tar.gz`:

- with `cache-1` to `cache-3`, every node owns at least 200 of the keys;
- after `Add("cache-4")`, between 150 and 350 keys move, and every key that
  moves, moves **to** `cache-4`.

**Think it through:**
- Which position does `slices.BinarySearch` give back when the hash is
  bigger than every point? What would `r.points[i]` do with it?
- Why check for an empty ring first, before you touch `r.points[0]`?

### 3. Remove a node's points

**Goal:** make `Remove(node)` delete every point owned by `node` from
`r.points`.

- `r.points` keeps only the points whose owner isn't `node`, still sorted.
- Removing a node that isn't on the ring does nothing.
- Removing the last node leaves no points, so `Get` returns `""`.

Leave `r.owners` alone in this task: the next one cleans it up.

```go
r := New(100)
r.Add("cache-1", "cache-2", "cache-3", "cache-4") // 400 points
r.Remove("cache-2")                               // 300 points left
// Keys that were on cache-2 now go to other nodes. Every other key stays.
```

**Think it through:**
- `r.points` holds numbers, not names. Which of the two fields tells you who
  owns a point?
- Does `slices.DeleteFunc` change the order of the points it keeps?

### 4. Forget a node's owners

**Goal:** make `Remove(node)` also delete every entry for `node` from
`r.owners`.

Without this, the removed node's points stay "taken" in `r.owners`, so
adding the node back skips every one of them and it gets no keys at all.

- After `Remove("cache-2")` on the 400-point ring, `len(r.owners)` is 300.
- Adding `cache-2` back sends every key to where it was before the
  `Remove`.
- Removing every node leaves `r.owners` empty.

Add `"maps"` to the imports, and put this code **after** your task 3 code.

```go
r.Remove("cache-2")
len(r.owners)    // => 300
r.Add("cache-2") // every key goes back to where it was before
```

**Think it through:**
- Your task 3 code asks `r.owners` who owns each point. What would it find
  if this code ran first?
- The function you pass gets a point and its owner. Which one do you check?

## Learn more

- [Wikipedia: Consistent hashing](https://en.wikipedia.org/wiki/Consistent_hashing)
- [pkg.go.dev: slices.DeleteFunc](https://pkg.go.dev/slices#DeleteFunc)
- [pkg.go.dev: maps.DeleteFunc](https://pkg.go.dev/maps#DeleteFunc)
- [groupcache's consistent hash ring](https://github.com/golang/groupcache/blob/master/consistenthash/consistenthash.go): the same idea in about 80 lines of real Go
