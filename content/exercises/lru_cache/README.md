# LRU caches: evict the least recently used

A cache with no size limit grows until the kernel kills your process. You'll
build an *LRU cache*: when it's full, it throws away the entry nobody has
used for the longest time, and every operation stays fast however big it is.

**New here:** bounded caches, eviction policies, doubly linked lists, `container/list`, O(1).

## Learn

### Why bound a cache?

A Kubernetes pod usually runs with a memory limit, such as `memory: 256Mi`.
Go over it and the kernel's *OOM killer* stops the process: the pod restarts
as `OOMKilled`, and the cache starts empty again, so every request hits the
slow backend at once.

The TTL cache from *TTL caching: expire on read* forgets old entries, but
only after their TTL. A burst of new keys can fill memory long before that. A
*bounded* cache holds at most `capacity` entries. Adding one more means
*evicting* (throwing away) one that's already there.

### Which entry goes?

The rule that picks the victim is the *eviction policy*:

| Policy | Evicts | Good at | Weak at |
|---|---|---|---|
| LRU, least recently used | the entry nobody has read or written for the longest | most workloads: what was used recently gets used again | one big scan of many keys pushes out everything else |
| LFU, least frequently used | the entry with the fewest hits | a stable set of favourites | needs a counter per entry; old favourites linger |
| FIFO, first in, first out | the oldest insert, even if it's busy | the simplest to build | throws out busy keys |
| Random | any entry | no bookkeeping at all | can throw out the busiest key |

LRU is the usual default. You'll build it here.

### A map plus a doubly linked list

LRU needs two things to be fast: find an entry by key, and keep entries in
the order they were used. No single structure does both, so you use two:

```text
items (map):    "api"    "cache"    "db"
                  │         │         │
                  ▼         ▼         ▼
order (list):   [api] ⇄  [cache] ⇄  [db]
                front               back
```

The front of the list is the most recently used entry, the back the least.

- The *map* takes a key straight to its element in the list.
- The *doubly linked list* keeps the order. Each element points to the one
  before it and the one after it, so you can unlink any element, or move it
  to the front, by changing a few pointers.

*O(1)* ("constant time") means an operation takes the same work whether the
cache holds 10 entries or 10 million. With a slice instead of a list, moving
an entry to the front shifts every entry before it: *O(n)*, slower as it grows.

| Operation | Map + list | Slice only |
|---|---|---|
| find a key | O(1): map lookup | O(n): scan |
| mark it most recent | O(1): move to front | O(n): shift entries |
| evict the least recent | O(1): remove the back | O(1): drop the last |

### container/list

The standard library's `container/list` is a doubly linked list. Every
element is a `*list.Element` with a `Value` field of type `any`:

```go
import "container/list"

recent := list.New()
first := recent.PushFront("deploy-41") // returns the new *list.Element
recent.PushFront("deploy-42")          // order: deploy-42, deploy-41
recent.MoveToFront(first)              // order: deploy-41, deploy-42
last := recent.Back()                  // the element holding "deploy-42"
recent.Remove(last)                    // order: deploy-41
recent.Len()                           // => 1
for el := recent.Front(); el != nil; el = el.Next() {
	fmt.Println(el.Value) // deploy-41
}
```

`Back()` returns `nil` on an empty list. `PushFront`'s result is worth
keeping: it's how you find that element again later without a scan.

> **Why is it written like this?** `container/list` is a package of the
> standard library, so Go requires the import. The `for el := ...; el !=
> nil; el = el.Next()` loop is the convention for walking a list. Using the
> package instead of writing your own pointers is the author's choice.

### Getting the entry back out

`Value` is `any`, so, as in *Type assertions and switches*, you assert the
type you put in:

```go
b := el.Value.(*build) // panics if Value is not a *build
b.status = "done"      // changes the build stored in the list
```

In this cache every element holds an `*entry` (a pointer), so the assertion
always succeeds, and changing a field through it changes the stored entry.

An `entry` stores its own `key` as well as its `value`. Evicting starts from
`c.order.Back()`, which gives you an element, not a key, and you need the
key to delete it from `c.items` too.

> **Why is it written like this?** The `x.(T)` syntax is what Go requires.
> Using the one-result form (which panics on the wrong type) instead of `v,
> ok := x.(T)` is the author's choice: nothing but `*entry` ever goes in the
> list. Storing a pointer is also a choice: with a plain `entry` you'd get a
> copy, and changing it would change nothing in the list.

### What's already written

`New`, `Len` (it counts `c.items`), `Keys` (it walks `c.order` from front to
back), `Evictions` (it returns `c.evictions`), and the locks: every method
locks `c.mu` and defers the unlock, as in *Mutexes*. The tests check `Keys`
and `Len`, so they see both the list and the map.

## Your task

Your team runs a container registry proxy. It caches which digest each image
tag points to (`nginx:1.27` → `sha256:4f1c`), and its pod has a memory
limit. Finish the LRU cache so it never holds more than `capacity` tags.

### 1. Add new entries

**Goal:** make `Put` add a new key as the most recently used entry.

For a key that isn't cached yet, `Put(key, value)` creates an `*entry`
holding both `key` and `value`, pushes it onto the **front** of `c.order`, and
stores the element it gets back in `c.items[key]`. Ignore existing keys and
the capacity for now: tasks 3 and 4 handle them.

```go
c := lru.New(3)
c.Put("nginx:1.27", "sha256:4f1c")
c.Put("redis:7.4", "sha256:9a0b")
c.Keys() // => ["redis:7.4" "nginx:1.27"]
c.Len()  // => 2
```

**Think it through:**
- Which `list` method adds at the front, and what does it return?
- The map holds `*list.Element` values. Where does that element come from?

### 2. Look up an entry and mark it used

**Goal:** make `Get` return a cached value and move that key to the front.

- Missing key: return `""` and `false`, and change nothing.
- Cached key: move its element to the front of `c.order`, then return the
  entry's `value` and `true`.

```go
// Keys() is ["postgres:16" "redis:7.4" "nginx:1.27"]
c.Get("nginx:1.27") // => "sha256:4f1c", true
c.Keys()            // => ["nginx:1.27" "postgres:16" "redis:7.4"]
c.Get("mysql:8")    // => "", false
```

**Think it through:**
- Use the two-value map lookup to tell a hit from a miss.
- The element's `Value` is `any`. How do you get the `*entry` out of it?

### 3. Update existing entries

**Goal:** make `Put` on a key that's already cached replace its value and move it to the front.

Change the `value` in the existing entry, move its element to the front of
`c.order`, and return. Don't add a second element: `Len()` stays the same.

```go
// Keys() is ["postgres:16" "redis:7.4" "nginx:1.27"]
c.Put("redis:7.4", "sha256:c41f")
c.Keys()           // => ["redis:7.4" "postgres:16" "nginx:1.27"]
c.Len()            // => 3
c.Get("redis:7.4") // => "sha256:c41f", true
```

**Think it through:**
- This check goes before the code you wrote in task 1. Why does the order matter?
- You change the entry through the pointer the assertion gives you.

### 4. Evict the least recently used entry

**Goal:** make `Put` evict the least recently used entry when a new key takes the cache over `c.capacity`.

After adding a **new** key, if the cache holds more than `c.capacity`
entries, take the element at the **back** of `c.order`, remove it from the
list, delete its key from `c.items`, and add 1 to `c.evictions`. Updating an
existing key (task 3) never evicts.

```go
c := lru.New(2)
c.Put("nginx:1.27", "sha256:4f1c")
c.Put("redis:7.4", "sha256:9a0b")
c.Get("nginx:1.27")                 // nginx is now the most recent
c.Put("postgres:16", "sha256:e83d") // evicts redis:7.4
c.Keys()                            // => ["postgres:16" "nginx:1.27"]
c.Evictions()                       // => 1
c.Get("redis:7.4")                  // => "", false
```

**Think it through:**
- Which end of the list holds the least recently used entry?
- The map needs the evicted key. Which field of the entry holds it?
- Use either `len(c.items)` or `c.order.Len()`: they stay equal.

## Learn more

- [pkg.go.dev: container/list](https://pkg.go.dev/container/list)
- [Wikipedia: Cache replacement policies](https://en.wikipedia.org/wiki/Cache_replacement_policies)
- [Wikipedia: Doubly linked list](https://en.wikipedia.org/wiki/Doubly_linked_list)
- [Kubernetes: Resource Management for Pods and Containers](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
