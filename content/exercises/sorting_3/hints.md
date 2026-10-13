## Task 1: counting

Make an empty map with `make(map[string]int)`, range over `firing`, and do
`counts[service]++` for each one. Missing keys start at `0`.

## Task 2: three steps

1. Get the names in a slice and sort them: most alerts first, then by name.
2. If `n` is smaller than the number of names, keep only `names[:n]`.
3. Loop over what's left and `append` a `fmt.Sprintf("%s: %d", ...)` line for
   each one.

Handle `n <= 0` before anything else.

## Task 2: the sort

```go
names := slices.Sorted(maps.Keys(counts))
slices.SortFunc(names, func(a, b string) int {
	if c := cmp.Compare(counts[b], counts[a]); c != 0 {
		return c
	}
	return cmp.Compare(a, b)
})
```
