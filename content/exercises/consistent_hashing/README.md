# Consistent hashing: a ring with virtual nodes

*Consistent hashing* spreads keys across servers so that adding or removing
one server moves only a small share of the keys. Cache clients, sharded
databases and load balancers (Envoy's "ring hash") all use it. You'll build
the hash ring at its heart.

**New here:** `hash/crc32`, `uint32`, the hash ring, virtual nodes, `slices.BinarySearch`, `slices.DeleteFunc`, `maps.DeleteFunc`.

## Learn

### Recap: why hash mod N hurts

In *Sharding: route keys and plan a rebalance* you picked a shard with
`hash % N`. Going from 3 shards to 4 changes the answer for about 3 keys in
4. For a cache, that means most lookups miss at once, right when you added
capacity.

### Hashes on a circle

`crc32.ChecksumIEEE` (from the `hash/crc32` package) turns bytes into a
`uint32`: a whole number from 0 to 4,294,967,295 (2³² − 1). The same input
always gives the same number:

```go
crc32.ChecksumIEEE([]byte("hello")) // => 907060870
```

The starter wraps it in a small `hash(s string) uint32` function.

Now bend that number line into a circle, so that after the largest number
comes 0 again. Nodes (servers) get hashed onto the circle, and so do keys.
Here is a toy ring with numbers 0–99 and three database nodes:

```text
 0 ─────── 10 ─────────── 40 ──────────── 70 ─────────── 99 ─┐
           db-a           db-b            db-c               │
 ▲                                                           │
 └──────────────────── wraps around to 0 ────────────────────┘
```

**The rule:** a key belongs to the first node at or after its hash, going
"clockwise" (to the right).

| key hash | owner | why |
|---|---|---|
| 25 | db-b | the first node at or after 25 is at 40 |
| 40 | db-b | exactly on a node counts |
| 85 | db-a | nothing after 85, so wrap around to the first node, at 10 |

### Why only a few keys move

Add `db-d` at 55. Only keys from 41 to 55 change owner: they were `db-c`'s
and are now `db-d`'s. Every other key stays put. Remove `db-b` and only its
keys (11 to 40) move, to the next node clockwise. With N nodes, a change
moves about 1/N of the keys instead of most of them.

### Virtual nodes

With one point per node, the gaps between points are uneven, so one node
can own most of the circle. Similar names make it worse with crc32. Here is
how 1000 keys spread over `cache-1`, `cache-2` and `cache-3` with the hash
in this exercise:

| points per node | cache-1 | cache-2 | cache-3 |
|---|---|---|---|
| 1 | 4 | 4 | 992 |
| 10 | 786 | 166 | 48 |
| 100 | 372 | 239 | 389 |

The fix is to put each node on the ring many times. Each point is a
*virtual node* (also called a *replica*). Each one is the hash of a label
made from the node name and a number: `db-a#0`, `db-a#1`, `db-a#2`, ... Many
small arcs average out.

**Why the separator matters.** Without the `#`, node `cache-1` replica 12
is `"cache-1" + "12"` = `cache-112`, and node `cache-11` replica 2 is
`"cache-11" + "2"` = `cache-112` too. Same label, same point: two nodes
fight over one spot. Hostnames never contain `#`, so with it every label is
unique.

> **Why is it written like this?** Go never turns an `int` into its digits
> by itself, so you need `strconv.Itoa(i)`. The `#` is the author's choice:
> any character that never appears in a node name works.

### Storing the ring: a sorted slice and a map

```go
points []uint32          // every point, sorted: 31, 97, 4000, ...
owners map[uint32]string // point -> node: 97 -> "db-b"
```

The sorted slice lets you find "the first point at or after this hash"
quickly. The map tells you whose point that is. `slices.Sort` sorts a slice
in place, as in *Sorting*.

### Binary search

`slices.BinarySearch(sorted, target)` returns two values: a position and
whether `target` is in the slice. The position is where `target` is, or
where it would go: the index of the **first element at or after** `target`.
When every element is smaller, the position is `len(sorted)`, one past the
end.

```go
ports := []int{22, 80, 443}
slices.BinarySearch(ports, 80)   // => 1, true
slices.BinarySearch(ports, 100)  // => 2, false (443 is the first one >= 100)
slices.BinarySearch(ports, 8080) // => 3, false (past the end: no such index!)
```

It halves the slice at every step, so 400 points take about 9 comparisons
instead of 400. `sort.Search` from the `sort` package does the same job in
older code.

> **Why is it written like this?** Binary search only works on a **sorted**
> slice. On an unsorted one it gives wrong answers, and no error tells you.
> Ignoring the second result with `_` is your choice: the ring only needs
> the position.

### Deleting by a condition

`slices.DeleteFunc` removes every element for which your function returns
`true`, keeping the rest in order. `maps.DeleteFunc` does the same for map
entries:

```go
hosts = slices.DeleteFunc(hosts, func(h string) bool {
	return strings.HasPrefix(h, "old-")
})
maps.DeleteFunc(ttls, func(host string, secs int) bool {
	return secs == 0
})
```

> **Why is it written like this?** `slices.DeleteFunc` returns the shorter
> slice, and Go requires you to store it (`hosts = ...`), or `hosts` keeps
> its old length. A map is changed in place, so `maps.DeleteFunc` returns
> nothing.

### No mutex this time

This `Ring` is not safe for concurrent use, to keep the focus on the ring
itself. That's the author's choice: a real one adds a `sync.RWMutex`, as in
*Mutexes*, with `Get` taking the read lock.

## Your task

Your CI runners keep build artifacts on three cache servers, `cache-1` to
`cache-3`. With `hash % 3`, adding `cache-4` sent three quarters of the
lookups to the wrong server, and every build slowed down. Build a hash ring,
so only about a quarter of the keys move.

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
- You need two loops, one inside the other. Which one gives you the node,
  and which one the number?
- How do you ask a map "do you have this key?" without caring about the
  value?
- Sorting once, at the end of `Add`, is enough. Why?

### 2. Find a key's node

**Goal:** make `Get(key)` return the node at the first point at or after
`hash(key)`, wrapping around after the last point.

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

### 3. Remove a node

**Goal:** make `Remove(node)` take every point of `node` off the ring.

- Delete every point owned by `node` from both `r.points` and `r.owners`.
- `r.points` stays sorted.
- Removing a node that isn't on the ring does nothing.
- Removing the last node leaves an empty ring, so `Get` returns `""`.

Add `"maps"` to the imports.

```go
r := New(100)
r.Add("cache-1", "cache-2", "cache-3", "cache-4") // 400 points
r.Remove("cache-2")                               // 300 points left
// Keys that were on cache-2 now go to other nodes. Every other key stays.
r.Add("cache-2") // every key goes back to where it was before
```

**Think it through:**
- Only `r.owners` knows who owns a point. So which of the two do you clean
  up first?
- Does `slices.DeleteFunc` change the order of the points it keeps?

## Learn more

- [Wikipedia: Consistent hashing](https://en.wikipedia.org/wiki/Consistent_hashing)
- [pkg.go.dev: slices.BinarySearch](https://pkg.go.dev/slices#BinarySearch)
- [pkg.go.dev: hash/crc32](https://pkg.go.dev/hash/crc32)
- [groupcache's consistent hash ring](https://github.com/golang/groupcache/blob/master/consistenthash/consistenthash.go): the same idea in about 80 lines of real Go
