# Slice gotchas: drain and promote without side effects

A load balancer keeps a pool of backend hosts. Your tooling drains hosts out
of the pool and promotes a replica to the front, and the caller's pool must
stay exactly as it was. In this exercise you write both helpers so they never
touch the slice they were given.

## Learn

### Recap

Slices share their backing array: a sub-slice, a returned slice, or the
result of an `append` with spare capacity can all write into the caller's
data. When a function must not change its input, build the result in a new
array: `make` + `copy`, `slices.Clone`, or a fresh slice you `append` to.

### slices.Delete changes the slice you pass in

`slices.Delete(s, i, j)` removes elements `i` to `j-1`. It does that *in
place*: it shifts the later elements left inside the same array and clears the
boxes at the end. The caller's slice still has its old length, so it now looks
broken:

```text
pool := []string{"web-1", "web-2", "web-3", "web-4"}
out := slices.Delete(pool, 1, 2)   // remove "web-2"

before:  | web-1 | web-2 | web-3 | web-4 |
after:   | web-1 | web-3 | web-4 |  ""   |
          <--------- out ------->
          <------------- pool ---------->
```

`out` is fine, but `pool` is now `["web-1" "web-3" "web-4" ""]`. If you
want to use it on a slice the caller still needs, clone first:
`slices.Delete(slices.Clone(pool), 1, 2)`. A loop that appends the hosts you
keep to a new slice works just as well.

### Two handy searches

```go
slices.Contains(pool, "web-3") // true if any element equals "web-3"
slices.Index(pool, "web-3")    // its position, or -1 if it isn't there
```

Add `"slices"` to the imports if you use any of these.

## Your task

Write two helpers for the load balancer pool. Neither may change `pool`, and
the slice each returns must be independent: changing it afterwards must not
change `pool`. The tests check both, with a `pool` that has spare capacity.

### 1. Drain a host

`Drain(pool, host)` returns the pool without `host`. Remove **every** copy of
`host`, and keep the other hosts in their original order. If `host` is not in
the pool, return a copy of the whole pool.

```go
pool := []string{"web-1", "web-2", "web-3", "web-2"}
Drain(pool, "web-2") // => ["web-1" "web-3"]
Drain(pool, "db-1")  // => ["web-1" "web-2" "web-3" "web-2"]
// pool is still ["web-1" "web-2" "web-3" "web-2"]
```

Draining the last host returns an empty slice (`nil` is fine).

### 2. Promote a replica

`Promote(pool, host)` returns the pool with `host` moved to the front and the
other hosts in their original order after it. In this task `host` appears at
most once in `pool`. If `host` is not in the pool, return a copy of the whole
pool, unchanged. `Drain` from task 1 is useful here.

```go
pool := []string{"db-1", "db-2", "db-3"}
Promote(pool, "db-3") // => ["db-3" "db-1" "db-2"]
Promote(pool, "db-1") // => ["db-1" "db-2" "db-3"]
Promote(pool, "db-9") // => ["db-1" "db-2" "db-3"]
// pool is still ["db-1" "db-2" "db-3"]
```

## Learn more

- [Package slices: Delete](https://pkg.go.dev/slices#Delete)
- [Package slices: Contains](https://pkg.go.dev/slices#Contains)
- [Go blog: Robust generic functions on slices](https://go.dev/blog/generic-slice-functions)
