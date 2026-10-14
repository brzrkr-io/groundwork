# Consistent hashing: fix the ring

A teammate wrote the hash ring that routes build-artifact keys to the cache
servers. Since it shipped, `web-11` sits nearly idle while `web-1` is
overloaded, some lookups get no server at all, and the ring routes keys to
the wrong server after a node is added. Find and fix the three bugs.

**New here:** what binary search does with an unsorted slice, `slices.IsSorted`.

## Learn

### Recap

As in *Consistent hashing: a ring with virtual nodes*, each node sits on a
circle of `uint32` numbers at `replicas` points. Each point is
`hash(label)`, and the label is the node name, `#` and the replica number.
`r.points` holds every point sorted from low to high, and `r.owners` maps a
point to its node. A key belongs to the node at the first point **at or
after** `hash(key)`, and a key past the last point wraps around to the
first point. Adding a node moves only about 1/N of the keys, instead of
most of them as with `hash % N` in *Sharding: route keys and plan a rebalance*.

### Labels must be unique

Two different labels that come out as the same text hash to the same point.
The ring skips points that are already taken, so whoever was added first
keeps the point and the other node silently loses it. Nothing crashes: one
node just gets less traffic than the others.

Joining two pieces of text without a separator is how that happens:

| first piece | second piece | joined |
|---|---|---|
| `"db-2"` | `"31"` | `"db-231"` |
| `"db-23"` | `"1"` | `"db-231"` |

A separator that never appears in either piece, like `#` or `/`, makes
every joined result unique: `db-2#31` and `db-23#1`.

### Binary search needs a sorted slice

`slices.BinarySearch` doesn't scan the slice. It looks at the middle
element, decides which half the target must be in, and throws the other
half away. That only works when the slice is sorted:

```go
ids := []int{40, 70, 10}       // not sorted
slices.BinarySearch(ids, 10)   // => 0, false (10 is right there, at index 2!)
slices.Sort(ids)               // [10 40 70]
slices.BinarySearch(ids, 10)   // => 0, true
```

There is no error and no panic: just a wrong position. `slices.IsSorted`
reports whether a slice is sorted, which is handy in tests.

> **Why is it written like this?** Go's `slices.BinarySearch` trusts you:
> checking that the slice is sorted would mean reading every element, which
> is exactly the work binary search exists to avoid. Keeping the slice
> sorted is your job.

### Two lists sorted on their own aren't sorted together

Sorting each batch on its own and then appending one after the other does
**not** give a sorted slice:

```go
batch1 := []int{10, 70} // sorted
batch2 := []int{40}     // sorted
all := append(batch1, batch2...)
// all = [10 70 40]: not sorted
```

## Your task

Fix the ring so that every node gets its fair share, every key gets a
node, and the answers don't depend on how many `Add` calls built the ring.
Fix them in order: the examples in task 2 and 3 assume the earlier fixes.

### 1. Fix the labels

**Goal:** make `label(node, i)` return the node name, then `#`, then `i`.

The buggy `label` joins the node name and the number with nothing in
between. Then `web-1` replica 12 and `web-11` replica 2 are both
`"web-112"`. With 20 replicas, 10 of `web-11`'s points are already taken by
`web-1`, so `web-11` only gets 10 points:

```go
label("web-1", 12)  // buggy: "web-112"   fixed: "web-1#12"
label("web-11", 2)  // buggy: "web-112"   fixed: "web-11#2"

r := New(20)
r.Add("web-1", "web-11")
// points owned:            buggy: web-1 20, web-11 10   fixed: 20 and 20
// share of 1000 keys:      buggy: web-1 813, web-11 187 fixed: 468 and 532
```

The tests also try `cache-1` and `cache-12` with 30 replicas: both must own
30 points.

**Think it through:**
- Which piece of fixed text goes between the two parameters?
- Look at the doc comment above `label`: it already says what the label
  should be.

### 2. Fix the wrap-around

**Goal:** make `Get` return the first point's node for a key that hashes
past the last point.

The buggy `Get` returns `""` when `slices.BinarySearch` gives back
`len(r.points)`. On a ring there is no "past the end": after the last point
comes the first one again. Only an empty ring returns `""`.

With `New(1)` and `Add("cache-1", "cache-2", "cache-3")` the points are
2089304447 (`cache-3`), 2102012744 (`cache-2`) and 2131531025 (`cache-1`):

```go
r.Get("build-127.tar.gz")  // hash 2097901471   buggy: "cache-2"  fixed: "cache-2"
r.Get("web/v2.0.0.tar.gz") // hash 3969333627   buggy: ""         fixed: "cache-3"
r.Get("build-65.tar.gz")   // hash 4292568909   buggy: ""         fixed: "cache-3"
New(100).Get("build-65.tar.gz") // empty ring: "" (both)
```

The tests also check that none of the keys `build-0.tar.gz` to
`build-999.tar.gz` gets `""` from a ring with three nodes and 100
replicas.

**Think it through:**
- Which position means "past the last point", and which position should
  it turn into?
- Once you index `r.points` for every non-empty ring, what does an empty
  ring need, and where does that check go?

### 3. Fix the sort

**Goal:** make `Add` leave the **whole** of `r.points` sorted, however many
times it is called.

The buggy `Add` collects the new points, sorts only those, and appends them
to `r.points`. The first `Add` is fine; the second one leaves the slice
unsorted, and binary search starts giving wrong answers:

```go
r := New(1)
r.Add("cache-1", "cache-2")
r.Add("cache-3")
r.points // buggy: [2102012744 2131531025 2089304447]
         // fixed: [2089304447 2102012744 2131531025]
r.Get("api/v1.4.2.tar.gz") // buggy: "cache-2"   fixed: "cache-3"
r.Get("web/v2.0.0.tar.gz") // buggy: "cache-2"   fixed: "cache-3"
```

The tests also build the same three-node ring with one `Add` and with three,
and check that `Get` gives the same answer for all 1000 `build-N.tar.gz`
keys.

**Think it through:**
- Which slice needs to be sorted when `Add` returns: the new points, or all
  of them?
- Do you still need the `added` slice once you append to `r.points`
  directly?

## Learn more

- [pkg.go.dev: slices.BinarySearch](https://pkg.go.dev/slices#BinarySearch)
- [Wikipedia: Binary search](https://en.wikipedia.org/wiki/Binary_search)
- [Wikipedia: Consistent hashing](https://en.wikipedia.org/wiki/Consistent_hashing)
