# append and copy: fix the inventory

Two small bugs in a host inventory tool: one list comes back with blank
entries, and the backup list comes back empty. Both are classic `append` and
`copy` mistakes you will meet in real code.

## Learn

### Recap

`s = append(s, v)` adds `v` after the last element and returns the longer
slice. `copy(dst, src)` copies the smaller of `len(dst)` and `len(src)`
elements and returns that number. `make([]T, n)` creates a slice that already
holds `n` zero values (`""` for strings).

### Length versus capacity in make

`make` takes an optional third argument, the capacity:

```go
a := make([]string, 3)    // len 3: ["" "" ""], append adds after them
b := make([]string, 0, 3) // len 0: [], with room for 3 before growing
```

Use `make([]T, 0, n)` (or just `var s []T`) when you plan to `append`. Use
`make([]T, n)` when you plan to fill it by index or with `copy`.

## Your task

The inventory tool prints connection addresses and saves a backup of the host
list before each change. Fix the bugs; don't rewrite the functions.

### 1. Fix the blank addresses

`Addresses(hosts, port)` should return each host followed by `:` and the port,
in the same order, with nothing else in the slice. Right now the result starts
with empty strings. If `hosts` is empty, return a slice with no elements.

```go
Addresses([]string{"web-1", "web-2"}, 8080)
// => ["web-1:8080" "web-2:8080"]

Addresses([]string{"db-1"}, 5432)
// => ["db-1:5432"]
```

### 2. Fix the empty backup

`Backup(hosts)` should return an independent copy of `hosts`: the same names in
the same order, unaffected by later changes to `hosts`. Right now it returns an
empty slice. Keep using `make` and `copy`.

```go
Backup([]string{"web-1", "web-2", "db-1"})
// => ["web-1" "web-2" "db-1"]
```

## Learn more

- [Go blog: Go Slices: usage and internals](https://go.dev/blog/slices-intro)
- [Go spec: Making slices, maps and channels](https://go.dev/ref/spec#Making_slices_maps_and_channels)
- [Go spec: Appending to and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices)
