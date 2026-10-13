# Slice gotchas: shared memory

A slice doesn't hold its elements. It points at an array somewhere else, and
other slices may point at the same array. That's what makes slices fast, and
it's behind some of the most confusing bugs in Go tools: a host list that
changes "by itself", or a tiny result that keeps a huge log in memory.

## Learn

### A slice is a small header

A slice value is three things, called the *slice header*: a pointer to the
first element it can see, a length (`len`) and a capacity (`cap`). The
elements live in a *backing array*:

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

`hosts[:2]` makes a new header pointing into the **same** array. Its capacity
runs to the end of that array:

```text
first := hosts[:2]

hosts = {ptr, len 4, cap 4}
first = {ptr, len 2, cap 4}     both point at web-1
          |
          v
        +-------+-------+-------+-------+
array:  | web-1 | web-2 | web-3 | web-4 |
        +-------+-------+-------+-------+
         <-- first -->
         <------------ hosts ------------>
```

So `first[0] = "web-9"` also changes `hosts[0]`. There is only one `web-1`
box.

### append can overwrite a neighbour

If a slice has spare capacity, `append` writes into the next box of the
array. It does **not** check whether another slice is using that box:

```go
first := hosts[:2]               // len 2, cap 4
first = append(first, "db-1")    // fits in cap: writes array[2]
// hosts is now ["web-1" "web-2" "db-1" "web-4"]
```

If there is no spare capacity (`len == cap`), `append` allocates a new, bigger
array, copies the elements over and writes there. The result no longer shares
anything with the original. You usually can't see from the code which of the
two happens, so don't rely on either.

### A small slice keeps the big array alive

Go frees an array only when nothing points into it any more. If you read a
100 000-line log and keep `lines[len(lines)-5:]`, that 5-element slice still
points into the 100 000-element array, so the whole array stays in memory.

### Three fixes

| Fix | What you get |
|---|---|
| `slices.Clone(s)` | A new array with the same elements. Its `cap` may be a little bigger than its `len`. |
| `make` + `copy` | A new array with exactly the `len` and `cap` you ask for. |
| `s[low:high:max]` | Same array, but `cap` is limited to `max - low`. |

The last one is the *full slice expression*. Setting `max` equal to `high`
gives a slice with no spare capacity, so the next `append` **must** copy
instead of overwriting a neighbour:

```go
first := hosts[0:2:2]         // len 2, cap 2
first = append(first, "db-1") // no room: new array
// hosts is still ["web-1" "web-2" "web-3" "web-4"]
```

`s[:len(s):len(s)]` means "all of `s`, with no spare capacity".

## Your task

Your rollout tool passes host lists and log lines between helpers. Three of
those helpers hand back slices that share memory with their input. Make each
one safe.

### 1. Copy the canary hosts

`Canaries(hosts, n)` returns the first `n` hosts (`n` is between 0 and
`len(hosts)`). The result must be an independent copy: changing it must not
change `hosts`. The test sets the result's first element to `"changed"` and
checks that `hosts` is untouched. Add `"slices"` to the imports if you use
`slices.Clone`.

```go
hosts := []string{"web-1", "web-2", "web-3", "web-4"}
c := Canaries(hosts, 2) // => ["web-1" "web-2"]
c[0] = "changed"
// hosts is still ["web-1" "web-2" "web-3" "web-4"]
```

### 2. Add a step without overwriting

`WithExtra(base, extra)` returns `base` with `extra` added at the end. It must
never write into `base`'s spare capacity, so two calls with the same `base`
don't overwrite each other. The test calls it twice with a `base` that has
spare capacity:

```go
base := make([]string, 2, 10) // len 2, cap 10
base[0], base[1] = "api", "worker"

a := WithExtra(base, "cron")
b := WithExtra(base, "metrics")
// a => ["api" "worker" "cron"]     (must not turn into "metrics")
// b => ["api" "worker" "metrics"]
```

The test also checks that the box right after `base` in its array is still
empty (`""`). Fix it with a full slice expression or with `slices.Clone`.

### 3. Keep only the tail of a log

`Tail(lines, n)` returns the last `n` lines, in order. If `n` is bigger than
`len(lines)`, it returns all of them. `n` is never negative.

The result must not point into `lines`, so the big log can be freed: the test
changes `lines` afterwards and checks your result did not change. Its `len`
**and** `cap` must both equal the number of lines returned. Use `make` with
that length and then `copy`. (`slices.Clone` may round the capacity up, so
don't rely on it here.)

```go
lines := []string{"boot", "pull image", "start", "ready", "serving"}
Tail(lines, 2)  // => ["ready" "serving"], len 2, cap 2
Tail(lines, 10) // => all 5 lines, len 5, cap 5
Tail(lines, 0)  // => [], len 0, cap 0
```

## Learn more

- [Go blog: Go Slices: usage and internals](https://go.dev/blog/slices-intro)
- [Go blog: Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
- [Go spec: Full slice expressions](https://go.dev/ref/spec#Slice_expressions)
- [Package slices: Clone](https://pkg.go.dev/slices#Clone)
