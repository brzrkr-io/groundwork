## Task 1: Count alerts per service

Make an empty map, then go through `firing` one name at a time and add one to
that name's count. A key that isn't in the map yet reads as `0`.

### Closer

Create the map with `make(map[string]int)`, loop with
`for _, service := range firing`, and use `++` on `counts[service]`.

### Answer

```go
counts := make(map[string]int)
for _, service := range firing {
	counts[service]++
}
return counts
```

`counts[service]++` starts missing keys at `0`, so the first sighting becomes `1`.

## Task 2: Report the top N

Work in three steps: get the service names into a sorted slice (most alerts
first, ties by name), keep only the first `n`, then turn each name into a
`"name: count"` line. Handle `n <= 0` before anything else.

### Closer

`slices.Sorted(maps.Keys(counts))` gives the names. In `slices.SortFunc`,
compare `counts[b]` with `counts[a]` (swapped, for biggest first), and only if
that's `0` compare `a` with `b`. Shorten with `names = names[:n]` only when
`n < len(names)`, and build lines with `fmt.Sprintf("%s: %d", ...)`.

### Answer

```go
if n <= 0 {
	return nil
}
names := slices.Sorted(maps.Keys(counts))
slices.SortFunc(names, func(a, b string) int {
	if c := cmp.Compare(counts[b], counts[a]); c != 0 {
		return c
	}
	return cmp.Compare(a, b)
})
if n < len(names) {
	names = names[:n]
}

var lines []string
for _, name := range names {
	lines = append(lines, fmt.Sprintf("%s: %d", name, counts[name]))
}
return lines
```

Swapping `a` and `b` in the first `cmp.Compare` sorts biggest first; the name
compare keeps ties in a fixed order.
