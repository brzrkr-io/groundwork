## Task 1: loop over the bounds

Loop with `for i, bound := range h.buckets`. The first time `v <= bound`,
add 1 to `h.counts[i]` and `return`. If the loop finishes without returning,
`v` is bigger than every bound: add 1 to `h.counts[len(h.buckets)]`.

## Task 2: keep a running total

Declare `var cumulative uint64` before the loop. In each step add
`h.counts[i]` to it, then print the line. The `+Inf` line is everything, which
is `h.count`.

## Task 3: subtract two times

`start := now()` before `next.ServeHTTP(w, r)`. After it,
`now().Sub(start)` is a `time.Duration`, and `.Seconds()` gives the float
`Observe` wants.

## Near-solution for task 2

```go
var cumulative uint64
for i, bound := range h.buckets {
	cumulative += h.counts[i]
	fmt.Fprintf(w, "%s_bucket{le=\"%s\"} %d\n", h.name, formatValue(bound), cumulative)
}
fmt.Fprintf(w, "%s_bucket{le=\"+Inf\"} %d\n", h.name, h.count)
```
