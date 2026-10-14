# Hash rings: find a key's node

A *hash ring* routes each key to one server so that adding a server moves
only a few keys, not most of them. You'll build a small one, one point per
server, and trace how keys find their owner.

**New here:** `hash/crc32`, the hash ring, `slices.BinarySearch`.

## Learn

### Recap: why hash mod N hurts

In *Sharding: route keys and plan a rebalance* you picked a shard with
`hash % n`. Going from 3 shards to 4 changes the answer for about 3 keys in
4. For a cache, that means most lookups miss at once, right when you added
capacity. A ring fixes that.

### crc32: another hash

`crc32.ChecksumIEEE` from the `hash/crc32` package is a hash like FNV-1a,
but it takes the whole `[]byte` in one call and returns the `uint32` right
away:

```go
crc32.ChecksumIEEE([]byte("hello")) // => 907060870
crc32.ChecksumIEEE([]byte("hello")) // => 907060870 (same input, same number)
```

A `uint32` goes from 0 to 4,294,967,295 (2³² − 1), as in *Sharding*.

> **Why is it written like this?** `ChecksumIEEE` wants a `[]byte`, so Go
> requires the conversion from `string`. "IEEE" names which CRC-32 table it
> uses; it's the common one. Picking crc32 over FNV-1a is the author's
> choice: real rings (groupcache, many cache clients) use it.

### Hashes on a circle

Bend that number line into a circle, so that after the largest number comes
0 again. Nodes (servers) get hashed onto the circle, and so do keys. Here is
a toy ring with numbers 0–99 and three database nodes:

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
and are now `db-d`'s. Every other key stays put. With `% n`, almost every
key would have moved.

### Storing the ring: a sorted slice and a map

```go
points []uint32          // every point, sorted: 10, 40, 70
owners map[uint32]string // point -> node: 40 -> "db-b"
```

The sorted slice lets you find "the first point at or after this hash"
quickly. The map tells you whose point that is. `slices.Sort` sorts a slice
in place, as in *Sorting*.

### Binary search

The *Algorithms and Big-O* guide showed binary search: look at the middle,
throw away the half that can't hold the target, repeat. 400 points take
about 9 looks instead of 400.

`slices.BinarySearch(sorted, target)` does it for you. It returns two
values: a position, and whether `target` is in the slice. The position is
where `target` is, or where it would go: the index of the **first element
at or after** `target`. When every element is smaller, the position is
`len(sorted)`, one past the end.

```go
ports := []int{22, 80, 443}
slices.BinarySearch(ports, 80)   // => 1, true
slices.BinarySearch(ports, 100)  // => 2, false (443 is the first one >= 100)
slices.BinarySearch(ports, 8080) // => 3, false (past the end: no such index!)
```

`sort.Search` from the `sort` package does the same job in older code.

> **Why is it written like this?** Binary search only works on a **sorted**
> slice. On an unsorted one it gives wrong answers, and no error tells you.
> Ignoring the second result with `_` is your choice when you only need the
> position.

## Your task

Your metrics pipeline sends each host's metrics to one of three collectors,
`collector-a` to `collector-c`. Build the ring that picks the collector, so
that adding `collector-d` later moves only some hosts.

### 1. Hash a name onto the ring

**Goal:** make `Point(s)` return the CRC-32 (IEEE) checksum of `s`.

Add `"hash/crc32"` to the imports.

```go
Point("collector-a") // => 3395146143
Point("web-02")      // => 1075636655
```

**Think it through:**
- Which function from `hash/crc32` returns a `uint32` in one call?
- What type does it want, and what type is `s`?

### 2. Find the next point

**Goal:** make `Next(points, h)` return the index of the first point at or
after `h`, or `0` when `h` is past the last point.

`points` is sorted from low to high and never empty. A hash exactly on a
point gives that point's index. Use the toy ring from Learn:

```go
toy := []uint32{10, 40, 70}
Next(toy, 25) // => 1 (40 is the first point at or after 25)
Next(toy, 40) // => 1 (exactly on a point)
Next(toy, 5)  // => 0
Next(toy, 85) // => 0 (past the last point: wrap around)
```

Add `"slices"` to the imports.

**Think it through:**
- Which position does `slices.BinarySearch` give back when `h` is bigger
  than every point? Is that a valid index?
- Which index is "wrapping around to the start"?

### 3. Place nodes on the ring

**Goal:** make `Add` put each node on the ring at the point `Point(node)`.

For every node:

- The point is `Point(node)`.
- If `r.owners` already has that point, skip it. That way, adding a node
  twice changes nothing.
- Otherwise record the node in `r.owners` and append the point to
  `r.points`.

When `Add` returns, `r.points` must be sorted from low to high, also after
the second, third, ... call.

```go
r := New()
r.Add("collector-a", "collector-b", "collector-c")
r.points             // => [609464499 1398046757 3395146143]
r.owners[1398046757] // => "collector-b"
r.Add("collector-b") // already on the ring: nothing changes
```

**Think it through:**
- How do you ask a map "do you have this key?" without caring about the
  value?
- Sorting once, at the end of `Add`, is enough. Why?

### 4. Find a key's node

**Goal:** make `Get(key)` return the node that owns `key`, or `""` if the
ring is empty.

- If the ring has no points, return `""`.
- Otherwise find the point with `Next` and `Point(key)`, and return its
  owner from `r.owners`.

With `collector-a` to `collector-c` on the ring:

| point | owner |
|---|---|
| 609464499 | collector-c |
| 1398046757 | collector-b |
| 3395146143 | collector-a |

```go
r.Get("web-23")      // hash   86883259 => "collector-c" (before the first point)
r.Get("web-02")      // hash 1075636655 => "collector-b"
r.Get("web-00")      // hash 2920450179 => "collector-a"
r.Get("collector-b") // hash 1398046757 => "collector-b" (exactly on a point)
r.Get("web-01")      // hash 3642079253 => "collector-c" (past the last point: wraps)
New().Get("web-01")  // => "" (empty ring)
```

The tests also check the whole idea with the 100 hosts `web-00` to
`web-99`: after `Add("collector-d")`, every host that changes collector
moves **to** `collector-d`.

**Think it through:**
- `Next` says `points` is never empty. Which check must come first?
- `Next` gives an index into `r.points`, not a point. How do you get the
  point, and then its owner?

## Learn more

- [pkg.go.dev: slices.BinarySearch](https://pkg.go.dev/slices#BinarySearch)
- [pkg.go.dev: hash/crc32](https://pkg.go.dev/hash/crc32)
- [Wikipedia: Consistent hashing](https://en.wikipedia.org/wiki/Consistent_hashing)
