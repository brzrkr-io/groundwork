# append and copy: fix the inventory

Two small bugs in a host inventory tool: one list comes back with blank
entries, and the backup comes back empty. Both are classic `append` and `copy`
mistakes you will meet in real code.

**New here:** the third argument of `make` (capacity).

## Learn

### Recap

| Call | What it does |
| --- | --- |
| `s = append(s, v)` | adds `v` **after** the last element |
| `copy(dst, src)` | copies the smaller of `len(dst)` and `len(src)` elements |
| `make([]T, n)` | a slice that already holds `n` zero values (`""` for strings) |

### Length versus capacity in make

`make` takes an optional third argument, the *capacity*: room reserved for
later, which doesn't count as elements yet.

```go
a := make([]int, 2)    // len 2: [0 0]; append adds after the zeros
b := make([]int, 0, 2) // len 0: [];   room for 2 before it must grow
```

| You plan to… | Make it with… |
| --- | --- |
| `append` | length `0` (`make([]T, 0, n)`, or just `var s []T`) |
| fill by index or with `copy` | length `n` (`make([]T, n)`) |

> **Why is it written like this?** Go requires the length; the capacity is
> optional. Passing a capacity when you know the final size is a convention
> that saves `append` from regrowing.

## Your task

The inventory tool prints connection addresses and saves a backup of the host
list before each change. In both functions the bug is on the `make` line
marked `TODO`. Fix that line; don't rewrite the functions.

### 1. Fix the blank addresses

**Goal:** make `Addresses(hosts, port)` return each host followed by `:` and
the port, with nothing else in the slice.

Same order as `hosts`. Right now the result starts with empty strings. If
`hosts` is empty or `nil`, return a slice with no elements.

```go
Addresses([]string{"web-1", "web-2"}, 8080)
// => ["web-1:8080" "web-2:8080"]

Addresses([]string{"db-1"}, 5432)
// => ["db-1:5432"]
```

**Think it through:**
- How many elements does `out` hold *before* the loop runs?
- `append` adds after those. Which argument of `make` sets that count?

### 2. Fix the empty backup

**Goal:** make `Backup(hosts)` return an independent copy of `hosts`.

Same names, same order, unaffected by later changes to `hosts`. Right now it
returns an empty slice. Keep using `make` and `copy`.

```go
Backup([]string{"web-1", "web-2", "db-1"})
// => ["web-1" "web-2" "db-1"]
```

**Think it through:**
- What is `len(backup)` right after `make`?
- `copy` only fills slots that exist. How many slots does it need?

## Learn more

- [Go blog: Go Slices: usage and internals](https://go.dev/blog/slices-intro)
- [Go spec: Making slices, maps and channels](https://go.dev/ref/spec#Making_slices_maps_and_channels)
- [Go spec: Appending to and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices)
