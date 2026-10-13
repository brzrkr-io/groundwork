# Looping with range: fix the alert math

An alerting script reports nonsense: the downtime total is tiny, and capping
latency spikes does nothing. Both bugs come from how `range` hands out its
loop variables.

**New here:** changing a slice *in place* inside a function.

## Learn

### Recap

| Loop | You get |
|---|---|
| `for i, v := range s` | index and a **copy** of the element |
| `for _, v := range s` | a copy of the element only |
| `for i := range s` | the **index** only, not the element |

Assigning to `v` never changes the slice. To change element `i`, assign to
`s[i]`.

> **Why is it written like this?** Go requires this order: the first variable
> is always the index. That's why a single variable is the index, whatever you
> name it.

### Changing a slice inside a function

A slice refers to its elements; it doesn't hold a private copy of them. So when
a function sets `s[i] = ...`, the caller sees the change:

```go
func Shout(words []string) {
	words[0] = "HEY" // the caller's slice changes too
}
```

That's why `CapLatencies` below returns nothing: it changes `ms` *in place*.

## Your task

Fix two bugs. Each needs only a small change to the `for` line or the line
inside the loop. Run the tests first and read what they report.

### 1. Fix the downtime total

**Goal:** make `TotalDowntime(outages)` return the total minutes, not a wrong
small number.

Each element of `outages` is one outage's length in minutes. An empty slice
gives 0.

```go
TotalDowntime([]int{30, 5, 12})
// => 47
```

Right now it returns 3 for that input.

**Think it through:**
- With one loop variable, what does `minutes` actually hold? Check the table.
- 3 is 0 + 1 + 2. Where do those numbers come from?

### 2. Fix the latency cap

**Goal:** make `CapLatencies(ms, limit)` change `ms` so no value is above
`limit`.

- Every value greater than `limit` becomes `limit`.
- Values at or below `limit` stay as they are.
- Return nothing; change `ms` itself.

```go
ms := []int{120, 5000, 80, 999}
CapLatencies(ms, 1000)
// ms is now [120 1000 80 999]
```

Right now `ms` comes back unchanged.

**Think it through:**
- `v = limit` runs. What does it change: the slice, or a copy?
- What do you need from the `for` line to write into the slice itself?

## Learn more

- [A Tour of Go: Range](https://go.dev/tour/moretypes/16)
- [Go spec: For statements with range clause](https://go.dev/ref/spec#For_range)
