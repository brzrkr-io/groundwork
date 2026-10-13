# Slice gotchas: shared memory

A slice doesn't hold its elements; it points at an array, and other slices may
point at the same one. That's behind confusing bugs in Go tools: a host list
that changes "by itself", or a tiny result that keeps a huge log in memory.

**New here:** slice headers, backing arrays, `slices.Clone`, `copy`, full slice expressions `s[low:high:max]`.

## Learn

### A slice is a small header

A slice value is three things, the *slice header*: a pointer to the first
element it can see, a length (`len`) and a capacity (`cap`). The elements live
in a *backing array*:

```text
hosts := []string{"web-1", "web-2", "web-3", "web-4"}

hosts = {ptr, len 4, cap 4}
          |
          v
        +-------+-------+-------+-------+
array:  | web-1 | web-2 | web-3 | web-4 |
        +-------+-------+-------+-------+
```

Copying a slice (`b := hosts`, passing it to a function, returning it) copies
only the header. The array is shared.

### Sub-slices share the array

`hosts[:2]` is a new header pointing into the **same** array. Its capacity
runs to the end of that array:

```text
first := hosts[:2]

hosts = {ptr, len 4, cap 4}
first = {ptr, len 2, cap 4}     both point at web-1
        +-------+-------+-------+-------+
array:  | web-1 | web-2 | web-3 | web-4 |
        +-------+-------+-------+-------+
         <-- first -->
         <------------ hosts ------------>
```

So `first[0] = "web-9"` also changes `hosts[0]`. There is only one box.

### `append` can overwrite a neighbour

If a slice has spare capacity, `append` writes into the next box of the
array, even if another slice uses that box:

```go
first := hosts[:2]            // len 2, cap 4
first = append(first, "db-1") // fits: writes array[2]
// hosts is now ["web-1" "web-2" "db-1" "web-4"]
```

If `len == cap`, `append` allocates a new, bigger array, copies over and
writes there. The result then shares nothing with the original.

### A small slice keeps the big array alive

Go frees an array only when nothing points into it. A 5-element slice taken
from the end of a 100 000-line log still points into that array, so all of it
stays in memory.

### Three fixes

| Fix | What you get |
|---|---|
| `slices.Clone(s)` | A new array with the same elements. `cap` may be a bit bigger than `len`. |
| `make([]T, n)` + `copy(dst, src)` | A new array with exactly `len` and `cap` `n`, then filled from `src`. |
| `s[low:high:max]` | Same array, but `cap` is limited to `max - low`. |

`copy(dst, src)` copies elements from `src` into `dst` (as many as fit) and
returns how many it copied:

```go
ports := make([]int, 2)        // [0 0], len 2, cap 2
copy(ports, []int{80, 443, 8080})
// ports is [80 443]
```

The third row is the *full slice expression*. When `max` equals `high`, there
is no spare capacity, so the next `append` **must** move to a new array:

```go
pair := hosts[1:3:3]          // ["web-2" "web-3"], len 2, cap 2
pair = append(pair, "db-1")   // no room: new array
// hosts is unchanged
```

> **Why is it written like this?** The colons and their order (`low:high:max`)
> are required by Go. Leaving out `low` means 0; `max` can never be left out
> in the three-part form. `slices.Clone` needs `"slices"` imported; `copy`,
> `make`, `len` and `cap` are built in, no import.

## Your task

Your rollout tool passes host lists and log lines between helpers. Three
helpers hand back slices that share memory with their input. Make each one
safe.

### 1. Copy the canary hosts

**Goal:** make `Canaries(hosts, n)` return an independent copy of the first `n` hosts.

`n` is between 0 and `len(hosts)`. The test sets the result's first element to
`"changed"` and checks that `hosts` is untouched.

```go
hosts := []string{"web-1", "web-2", "web-3", "web-4"}
c := Canaries(hosts, 2) // => ["web-1" "web-2"]
c[0] = "changed"
// hosts is still ["web-1" "web-2" "web-3" "web-4"]
```

**Think it through:**
- The starter's result is the right elements, but whose array is it in?
- Which fix from the table gives a new array? Remember its import.

### 2. Add a step without overwriting

**Goal:** make `WithExtra(base, extra)` return `base` plus `extra` at the end, without ever writing into `base`'s spare capacity.

The test calls it twice with the same `base`, which has spare capacity:

```go
base := make([]string, 2, 10) // len 2, cap 10
base[0], base[1] = "api", "worker"

a := WithExtra(base, "cron")
b := WithExtra(base, "metrics")
// a => ["api" "worker" "cron"]     (must not turn into "metrics")
// b => ["api" "worker" "metrics"]
```

It also checks that the box right after `base` in its array is still `""`.

**Think it through:**
- Why does the starter's `append` write into `base`'s array?
- What must be true of `len` and `cap` for `append` to move to a new array?
- Which fix in the table limits `cap` without copying? What values for
  `low`, `high` and `max` mean "all of `base`, no spare room"?

### 3. Keep only the tail of a log

**Goal:** make `Tail(lines, n)` return a copy of the last `n` lines, in order, with `len` and `cap` both equal to the number returned.

- If `n` is bigger than `len(lines)`, return all lines (the starter already
  handles this). `n` is never negative.
- The result must not point into `lines`: the test changes `lines` afterwards
  and checks your result didn't change.
- Use `make` + `copy`: `slices.Clone` may round the capacity up.

```go
lines := []string{"boot", "pull image", "start", "ready", "serving"}
Tail(lines, 2)  // => ["ready" "serving"], len 2, cap 2
Tail(lines, 10) // => all 5 lines, len 5, cap 5
Tail(lines, 0)  // => [], len 0, cap 0
```

**Think it through:**
- The starter already computes the right lines as a sub-slice. Keep that as
  your *source*.
- How big must the new slice from `make` be?
- `copy` takes destination first, then source.

## Learn more

- [Go blog: Go Slices: usage and internals](https://go.dev/blog/slices-intro)
- [Go blog: Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
- [Go spec: Full slice expressions](https://go.dev/ref/spec#Slice_expressions)
- [Package slices: Clone](https://pkg.go.dev/slices#Clone)
