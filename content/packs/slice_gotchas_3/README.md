# Slice gotchas: drain and promote without side effects

A load balancer keeps a pool of backend hosts. Your tooling drains hosts out
of the pool and promotes a replica to the front, and the caller's pool must
stay exactly as it was.

**New here:** `slices.Contains`, `slices.Index`, why `slices.Delete` changes
its input.

## Learn

### Recap: don't write into the caller's array

Slices share their backing array. A sub-slice, a returned slice, or an
`append` with spare room can all write into the caller's data. When a
function must not change its input, build the result in a **new** array.

Three ways to get a new array:

| Way | Example |
|---|---|
| copy everything | `slices.Clone(s)` |
| start empty and grow | `var out []int` then `out = append(out, x)` |
| start with one item | `[]int{x}`, then `append` more |

A slice that starts empty (or with a `[]T{...}` literal) owns its array, so
nothing you append to it can reach the input.

### slices.Delete changes the slice you pass in

`slices.Delete(s, i, j)` removes elements `i` to `j-1` *in place*: it shifts
later elements left inside the same array and blanks the boxes at the end.

```text
pool := []string{"web-1", "web-2", "web-3", "web-4"}
out := slices.Delete(pool, 1, 2)   // remove "web-2"

before:  | web-1 | web-2 | web-3 | web-4 |
after:   | web-1 | web-3 | web-4 |  ""   |
          <--------- out ------->
          <------------- pool ---------->
```

`out` is fine, but the caller's `pool` is now `["web-1" "web-3" "web-4" ""]`.
So don't use `Delete` on a slice the caller still needs, unless you clone it
first.

### Two handy searches

```go
ports := []int{80, 443, 8080}
slices.Contains(ports, 443) // true: some element equals 443
slices.Index(ports, 22)     // -1: 22 isn't there
```

### Appending a whole slice

Add `...` after a slice to append all its elements at once:

```go
a := []int{1}
b := []int{2, 3}
a = append(a, b...)   // [1 2 3]
```

> **Why is it written like this?** Go requires the `...`: `append`
> takes single items, and `...` spreads a slice into them. The `"slices"`
> import is required whenever you call `slices.Something`.

## Your task

Write two helpers for the pool. The rules for both:

- Never change `pool`.
- The returned slice must be independent: the test sets the result's `[0]`
  to `"changed"` and checks that `pool` didn't change.
- The test's `pool` has spare capacity, so sharing bugs will show up.

### 1. Drain a host

**Goal:** make `Drain(pool, host)` return a new slice with every copy of
`host` removed.

Keep the other hosts in their original order. If `host` isn't in the pool,
return a copy of the whole pool. Draining the only host returns an empty
slice (`nil` is fine).

```go
pool := []string{"web-1", "web-2", "web-3", "web-2"}
Drain(pool, "web-2") // => ["web-1" "web-3"]
Drain(pool, "db-1")  // => ["web-1" "web-2" "web-3" "web-2"]
// pool is still ["web-1" "web-2" "web-3" "web-2"]
```

**Think it through:**
- Instead of removing hosts, which hosts do you want to *keep*?
- Where will you put them so the result has its own array?

### 2. Promote a replica

**Goal:** make `Promote(pool, host)` return a new slice with `host` first and
the other hosts after it, in their original order.

`host` appears at most once. If `host` isn't in the pool, return a copy of
the pool, unchanged. If it's already first, the result looks the same as the
pool (but is still a new slice).

```go
pool := []string{"db-1", "db-2", "db-3"}
Promote(pool, "db-3") // => ["db-3" "db-1" "db-2"]
Promote(pool, "db-1") // => ["db-1" "db-2" "db-3"]
Promote(pool, "db-9") // => ["db-1" "db-2" "db-3"]
// pool is still ["db-1" "db-2" "db-3"]
```

**Think it through:**
- First: how do you find out whether `host` is in the pool at all?
- "Every host except `host`": which function from task 1 already gives you
  that?
- What new slice can you start with so `host` ends up first?

## Learn more

- [Package slices: Delete](https://pkg.go.dev/slices#Delete)
- [Package slices: Contains](https://pkg.go.dev/slices#Contains)
- [Go blog: Robust generic functions on slices](https://go.dev/blog/generic-slice-functions)
