# Slices: rolling-deploy batches

A *rolling deploy* updates a fleet a few hosts at a time, so most of the fleet
keeps serving traffic. Kubernetes Deployments and most CI/CD tools work this
way. You will write the part of a deploy tool that splits the host list into
batches.

## Learn

Recap from *Slices*: `hosts[a:b]` is the hosts from index `a` up to, but not
including, `b`, and both bounds must stay between `0` and `len(hosts)`.
`make([]T, n)` creates a slice of `n` zero values that you fill by index.

### New: a slice of slices

The elements of a slice can be anything, including other slices. `[][]string`
is a slice where each element is a `[]string`:

```go
plan := make([][]string, 2) // two batches, both nil for now
plan[0] = []string{"web-1", "web-2"}
plan[1] = []string{"web-3"}
len(plan)    // 2 (batches)
len(plan[0]) // 2 (hosts in the first batch)
plan[1][0]   // "web-3"
```

### New: dividing and rounding up

Integer division in Go rounds down: `5 / 2` is `2`. To round up instead, add
`size - 1` before dividing: `(5 + 2 - 1) / 2` is `3`.

The built-in `min(a, b)` returns the smaller of two numbers, which is handy
for keeping a slice bound from running past the end: `min(10, len(hosts))`.

## Your task

The deploy tool gets a list of hosts and a batch `size` (how many hosts to
update at once). `size` is always at least `1`. Batches are numbered from `0`.
Don't use `append` yet.

### 1. Count the batches

Change `BatchCount(hosts []string, size int) int` so it returns how many
batches are needed to cover every host. The last batch may be smaller than
`size`. No hosts means `0` batches.

```go
BatchCount([]string{"web-1", "web-2", "web-3", "web-4", "web-5"}, 2) // => 3
BatchCount([]string{"web-1", "web-2", "web-3", "web-4"}, 2)          // => 2
BatchCount([]string{}, 3)                                            // => 0
```

### 2. Pick one batch

Change `Batch(hosts []string, size, n int) []string` so it returns batch
number `n`: the hosts from index `n*size` up to `size` hosts later. The last
batch stops at the end of `hosts`, and a batch number past the end returns an
empty result (length `0`). `n` is never negative.

```go
hosts := []string{"web-1", "web-2", "web-3", "web-4", "web-5"}
Batch(hosts, 2, 0) // => []string{"web-1", "web-2"}
Batch(hosts, 2, 2) // => []string{"web-5"}
Batch(hosts, 2, 3) // => []string{}
```

### 3. Plan the whole rollout

Change `Plan(hosts []string, size int) [][]string` so it returns every batch,
in order. Use `make` with the number of batches, then fill each element. An
empty host list gives an empty plan.

```go
Plan([]string{"web-1", "web-2", "web-3", "web-4", "web-5"}, 2)
// => [][]string{{"web-1", "web-2"}, {"web-3", "web-4"}, {"web-5"}}

Plan([]string{"web-1", "web-2"}, 5)
// => [][]string{{"web-1", "web-2"}}
```

## Learn more

- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go spec: Slice expressions](https://go.dev/ref/spec#Slice_expressions)
- [Kubernetes: Rolling update deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment)
