# append and copy: rolling deploy plan

A rolling deploy never touches every host at once. It deploys the canaries
first, then the rest in small batches, so a bad release only hurts a few hosts.
Here you build that plan with `append` and `copy`.

## Learn

### Recap

Build results with `var out []T` and `out = append(out, v)` in a loop. Use
`append(a, b...)` to add all of `b`. Use `slices.Clone(s)` (import `"slices"`)
or `make` + `copy` when you need a copy that doesn't change with the original.

### A slice of slices

The element type of a slice can be another slice. `[][]string` is a slice whose
elements are `[]string` values. Each element is one batch of hosts:

```go
batches := [][]string{
	{"web-1", "web-2"}, // batches[0]
	{"web-3"},          // batches[1]
}
len(batches)    // 2
batches[0][1]   // "web-2"

var more [][]string
more = append(more, []string{"db-1", "db-2"}) // append one whole batch
```

## Your task

Your deploy tool reads two host lists from the inventory: canaries and regular
hosts. The inventory export sometimes contains blank names (`""`).

### 1. Build the deploy plan

`Plan(canaries, hosts)` returns one list: every canary in order, then every
host in order, skipping blank names. If there is nothing left, return a slice
with no elements.

```go
Plan([]string{"web-0"}, []string{"web-1", "", "web-2"})
// => ["web-0" "web-1" "web-2"]

Plan([]string{"", ""}, nil)
// => []
```

### 2. Split it into batches

`Batches(hosts, size)` splits `hosts` into batches of `size` hosts each, in
order. The last batch holds whatever is left, so it may be smaller.

- If `hosts` is empty, return no batches.
- If `size` is less than 1, return no batches.
- Each batch is its own copy: changing `hosts` afterwards must not change the
  batches. The test checks this by editing `hosts` after the call.

```go
Batches([]string{"web-1", "web-2", "web-3", "web-4", "web-5"}, 2)
// => [["web-1" "web-2"] ["web-3" "web-4"] ["web-5"]]

Batches([]string{"web-1", "web-2"}, 5)
// => [["web-1" "web-2"]]

Batches([]string{"web-1"}, 0)
// => []
```

A plain three-part loop that jumps by `size` (`start += size`) and the slice
expression `hosts[start:end]` from *Slices* do most of the work. Make sure
`end` never goes past `len(hosts)`: the built-in `min`, as in *For loops: plan
a rollout*, picks the smaller of `start + size` and `len(hosts)`.

## Learn more

- [Go by Example: Slices](https://gobyexample.com/slices)
- [Package slices](https://pkg.go.dev/slices)
- [Kubernetes: Performing a Rolling Update](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/)
