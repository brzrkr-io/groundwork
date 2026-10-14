## Task 1: Put each observation in its bucket

Walk the bounds from smallest to largest and stop at the first one that `v`
fits under. If none fits, the value goes in the overflow slot at the end.

### Closer

Loop with `for i, bound := range h.buckets`. When `v <= bound`, add 1 to
`h.counts[i]` and `return`. After the loop, use `h.counts[len(h.buckets)]`.

### Answer

```go
for i, bound := range h.buckets {
	if v <= bound {
		h.counts[i]++
		return
	}
}
h.counts[len(h.buckets)]++
```

## Task 2: Write cumulative buckets

Keep a running total while you loop over the buckets, and print the total
after adding each bucket's own count. The `+Inf` line counts everything.

### Closer

Declare `var cumulative uint64` before the loop and add `h.counts[i]` in each
step. Print the bound with `formatValue(bound)`. The `+Inf` count is
`h.count`. Write a literal `"` as `\"`: `"%s_bucket{le=\"%s\"} %d\n"`.

### Answer

```go
var cumulative uint64
for i, bound := range h.buckets {
	cumulative += h.counts[i]
	fmt.Fprintf(w, "%s_bucket{le=\"%s\"} %d\n", h.name, formatValue(bound), cumulative)
}
fmt.Fprintf(w, "%s_bucket{le=\"+Inf\"} %d\n", h.name, h.count)
```

## Task 3: Time every request

Read the clock with `now()` before calling `next`, and again after. The
difference is how long the request took.

### Closer

`start := now()` before `next.ServeHTTP(w, r)`. After it,
`now().Sub(start)` is a `time.Duration`, and `.Seconds()` gives the
`float64` that `Observe` wants.

### Answer

```go
start := now()
next.ServeHTTP(w, r)
h.Observe(now().Sub(start).Seconds())
```
