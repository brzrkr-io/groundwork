# LRU caches: fix the stale cache-aside reads

During last week's incident someone switched the `new-checkout` flag off in
the database, and the flag API kept answering `on` for another hour. The API
keeps an LRU cache in front of a slow database, and the code around that
cache has three bugs. Fix them.

**New here:** the cache-aside pattern, cache invalidation, the database as the source of truth.

## Learn

### Recap

As in *LRU caches: evict the least recently used*, an LRU cache is a map plus
a doubly linked list. `Get` marks a key as the most recently used, and `Put`
evicts the least recently used key when the cache is full.

`lru.go` holds a working one, `LRU`, with `NewLRU`, `Get` and `Put`, plus
one new method: `Remove(key)` deletes a key, and does nothing if the key
isn't cached. The `LRU` locks its own mutex, so `Service` needs no lock of
its own. You don't need to change `lru.go`.

### Cache-aside

In *cache-aside*, the cache never talks to the database. Your code sits
beside both and does the work. A read goes like this:

```text
Get(name)
  1. ask the cache      hit?   return the cached value
  2. ask the store      error? return the error, cache nothing
  3. put the value in the cache, then return it
```

Here the database is a `Store`: an interface with `Get`, `Set` and `Delete`.
When it has no flag by that name, `Get` returns `ErrNotFound`.

### The store is the source of truth

The *source of truth* is the one place that holds the real value. Here it's
the store. The cache only holds copies, and a copy can go *stale*: the store
changes, and the copy doesn't.

So the cache may only hold what the store actually returned. An error isn't
a value: if you cache `""` after a failed read, that flag looks empty until
the cache evicts it, even after the database recovers or someone creates the
flag. (Some systems do cache "not found" on purpose, for a few seconds. This
cache has no TTL, so it caches successful reads only.)

### Writes: store first, then invalidate

A write changes the store, then *invalidates* the cached copy (removes it).
The next `Get` misses and reads the new value from the store:

```text
Set(name, value)
  1. write the store    error? return the error; nothing has changed
  2. remove name from the cache
```

Why not write the cache first? If the store write then fails, the cache
serves a value the database never stored.

Why remove the copy instead of putting the new value in? Two writers can
finish in either order:

```text
writer A: store.Set("on") ......................... cache.Put("on")
writer B: ........ store.Set("off")  cache.Put("off")
result:   store holds "off", cache holds "on"   (stale until evicted)
```

If both remove the key instead, the order doesn't matter: the next `Get`
reads `"off"` from the store. A smaller window remains (a slow `Get` can put
an old value back just after a `Remove`), which is why real caches also give
entries a TTL. You don't need to handle it here.

> **Why is it written like this?** `Store` is an interface so the tests can
> pass in a fake database, as in *Interfaces: swap in a fake for tests*.
> Comparing errors with `errors.Is(err, ErrNotFound)` and returning store
> errors unchanged are Go conventions. Calling the copy-removing step
> "invalidate" is the usual name for it.

### Reading the tests

The tests use `fakeStore`, a `Store` kept in a map called `flags`:

- `reads` counts every call to its `Get`, so a test can tell a cache hit
  (no read) from a miss (one read).
- `failReads` and `failWrites` make it return `errStoreDown`
  (`"store unavailable"`), like a database that's down.
- A test that writes to `store.flags` directly plays another replica of the
  API, changing the database behind this one's back.

## Your task

The feature-flag API answers every request from `Service`. Make its cache
follow cache-aside, so a flag change in the database is never hidden by a
stale copy.

### 1. Cache only successful reads

**Goal:** make `Get` put a value in the cache only when the store returned it with a `nil` error.

`Get` caches whatever the store returned, even the `""` that comes with an
error:

```text
Get("new-checkout")   => "", ErrNotFound
(another replica creates new-checkout = "on")
Get("new-checkout")   => "on", nil    (buggy version: "", nil)
```

The rules:

- A cache hit returns the cached value and `nil`, without reading the store.
- On a miss, read the store. If that fails, return `""` and the store's error
  unchanged (`ErrNotFound` or any other), and cache nothing.
- Otherwise cache the value and return it with `nil`.

**Think it through:**
- Look at the order of the lines after the store read. What happens to `value` before anyone checks `err`?

### 2. Write the store first, then invalidate

**Goal:** make `Set` write the store first and then remove the flag from the cache, instead of caching the new value.

`Set` puts the new value in the cache before the store has saved it:

```text
Get("dark-mode")           => "off", nil   (now cached)
(the database goes down)
Set("dark-mode", "on")     => store unavailable
Get("dark-mode")           => "off", nil   (buggy version: "on", nil)
```

The rules:

- Write the store. If that fails, return the store's error and leave the
  cache alone.
- If it succeeds, remove `name` from the cache and return `nil`.
- `Set` never puts the new value in the cache: the next `Get` reads it from
  the store. The tests count store reads to check this.

**Think it through:**
- Which `LRU` method takes a key out of the cache?
- Where does the error check go, so a failed write never reaches the cache?

### 3. Invalidate on delete

**Goal:** make `Delete` remove the flag from the cache after deleting it from the store.

`Delete` removes the flag from the store, but the cache keeps serving it:

```text
Get("new-checkout")      => "on", nil   (now cached)
Delete("new-checkout")   => nil
Get("new-checkout")      => "", ErrNotFound   (buggy version: "on", nil)
```

The rules:

- Delete from the store. If that fails, return the store's error.
- If it succeeds, remove `name` from the cache and return `nil`. Other
  cached flags stay cached.

**Think it through:**
- This has the same shape as your fixed `Set`. Which line is different?

## Learn more

- [AWS: Caching strategies for ElastiCache](https://docs.aws.amazon.com/AmazonElastiCache/latest/dg/Strategies.html)
- [Microsoft Azure: Cache-Aside pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/cache-aside)
- [Wikipedia: Cache invalidation](https://en.wikipedia.org/wiki/Cache_invalidation)
