# Slices: rolling-deploy batches

A *rolling deploy* updates a fleet a few hosts at a time, so most of the fleet
keeps serving traffic. Kubernetes Deployments and most CI/CD tools work this
way. You will write the part of a deploy tool that splits the host list into
batches.

**New here:** slices of slices (`[][]string`), rounding up with integer
division, the built-in `min`.

## Learn

### Recap

`hosts[a:b]` is the hosts from index `a` up to, but not including, `b`. Both
bounds must stay between `0` and `len(hosts)`, or the program panics.
`make([]T, n)` creates a slice of `n` zero values that you fill by index.

### A slice of slices

A slice can hold anything, including other slices. `[][]string` is a slice
where each element is a `[]string`:

```go
shifts := make([][]string, 2) // two shifts, both nil for now
shifts[0] = []string{"ana", "bo"}
shifts[1] = []string{"cy"}
len(shifts)    // 2 (shifts)
len(shifts[0]) // 2 (people in the first shift)
shifts[1][0]   // "cy"
```

> **Why is it written like this?** Go requires one `[]` per level: `[]string`
> is a list of strings, `[][]string` a list of those lists. The names are the
> author's choice.

### Rounding up when dividing

Dividing two `int`s in Go drops the remainder: `7 / 3` is `2`. To round up
instead, add "divisor minus one" before dividing:

```go
pods, perNode := 7, 3
(pods + perNode - 1) / perNode // (7 + 2) / 3 = 3 nodes
```

The extra `perNode - 1` pushes any leftover into one more group, but is not
enough to add a group when there is no leftover (`6` pods still gives `2`).

### min keeps a bound in range

The built-in `min(a, b)` returns the smaller number. Use it to stop a bound
from running past the end:

```go
end := min(10, len(logs)) // 10, or fewer if logs is shorter
first := logs[:end]       // never panics
```

> **Why is it written like this?** `min` is built into Go (since Go 1.21), so
> it needs no import. Clamping bounds like this is a common convention.

## Your task

The deploy tool gets a list of hosts and a batch `size` (how many hosts to
update at once). `size` is always at least `1`. Batches are numbered from `0`.
Don't use `append` yet.

### 1. Count the batches

**Goal:** make `BatchCount(hosts, size)` return how many batches it takes to
cover every host.

The last batch may be smaller than `size`. No hosts (empty or `nil`) means `0`
batches. A `size` bigger than the host count means `1` batch.

```go
BatchCount([]string{"web-1", "web-2", "web-3", "web-4", "web-5"}, 2) // => 3
BatchCount([]string{"web-1", "web-2", "web-3", "web-4"}, 2)          // => 2
BatchCount([]string{}, 3)                                            // => 0
```

**Think it through:**
- Which number is divided, and by which? (`hosts` is a slice, not a number.)
- Does the result need rounding down or up?

### 2. Pick one batch

**Goal:** make `Batch(hosts, size, n)` return batch number `n`.

Batch `n` starts at index `n*size` and holds up to `size` hosts. The last batch
stops at the end of `hosts`. A batch number past the end returns an empty
slice (length `0`), not a panic. `n` is never negative.

```go
hosts := []string{"web-1", "web-2", "web-3", "web-4", "web-5"}
Batch(hosts, 2, 0) // => []string{"web-1", "web-2"}
Batch(hosts, 2, 2) // => []string{"web-5"}
Batch(hosts, 2, 3) // => []string{}
```

**Think it through:**
- Work out the start and end index first, each on its own line.
- Either bound can go past `len(hosts)`. What keeps each one in range?

### 3. Plan the whole rollout

**Goal:** make `Plan(hosts, size)` return every batch, in order, as a
`[][]string`.

Use `make` with the number of batches, then fill each element by index. An
empty host list gives an empty plan.

```go
Plan([]string{"web-1", "web-2", "web-3", "web-4", "web-5"}, 2)
// => [][]string{{"web-1", "web-2"}, {"web-3", "web-4"}, {"web-5"}}

Plan([]string{"web-1", "web-2"}, 5)
// => [][]string{{"web-1", "web-2"}}
```

**Think it through:**
- Which of your two functions says how long the plan is?
- Which one gives you the batch for each position?

## Learn more

- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go spec: Slice expressions](https://go.dev/ref/spec#Slice_expressions)
- [Go spec: Min and max](https://go.dev/ref/spec#Min_and_max)
- [Kubernetes: Rolling update deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment)
