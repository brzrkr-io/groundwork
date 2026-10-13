# Looping with range: fix the alert math

An alerting script reports nonsense: the downtime total is tiny, and capping
latency spikes does nothing. Both bugs come from how `range` hands out its
loop variables.

## Learn

### Recap

`for i, v := range s` gives the index `i` and a **copy** of the element, `v`.
`for _, v := range s` gives only the element. `for i := range s` gives only
the **index**, not the element. Assigning to `v` never changes the slice; to
change element `i`, assign to `s[i]`.

### Changing a slice inside a function

When a function receives a slice and sets `s[i] = ...`, the caller sees the
change. A slice refers to its elements instead of holding its own copy of
them, so both the function and the caller look at the same elements. That's
why `CapLatencies` below returns nothing: it changes `ms` *in place*.

## Your task

Fix the two bugs. Each one needs only a small change to the loop.

### 1. Fix the downtime total

`TotalDowntime(outages)` returns the total minutes of downtime, where each
element of `outages` is one outage's length in minutes. An empty slice gives 0.

```go
TotalDowntime([]int{30, 5, 12})
// => 47
```

Right now it returns 3 for that input.

### 2. Fix the latency cap

`CapLatencies(ms, limit)` changes `ms` in place: every value greater than
`limit` becomes `limit`. Values at or below `limit` stay as they are.

```go
ms := []int{120, 5000, 80, 999}
CapLatencies(ms, 1000)
// ms is now [120 1000 80 999]
```

Right now `ms` comes back unchanged.

## Learn more

- [A Tour of Go: Range](https://go.dev/tour/moretypes/16)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
