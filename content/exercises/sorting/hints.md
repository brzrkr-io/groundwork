## Task 1: Sort hosts without touching the input

`slices.Sort` rearranges the slice you give it, and `hosts` belongs to the
caller. Sort a copy instead.

### Closer

`slices.Clone(hosts)` makes a new slice with the same values. Sort that one
with `slices.Sort` and return it.

### Answer

```go
out := slices.Clone(hosts)
slices.Sort(out)
return out
```

Only `out` is reordered, so the caller's `hosts` keeps its order.

## Task 2: List service names in order

You need the map's keys, sorted, as a new slice. The `maps` and `slices`
packages can do this in one line together.

### Closer

`maps.Keys(errs)` hands out the keys, and `slices.Sorted(...)` collects them
into a sorted slice. Add `"maps"` to the `import` block.

### Answer

```go
return slices.Sorted(maps.Keys(errs))
```

`slices.Sorted` builds a new slice, so the map is unchanged.

## Task 3: Rank services by errors

The comparison function first compares error counts, but reversed so bigger
counts come first. Only when the counts are equal should it look at the names.

### Closer

`cmp.Compare(errs[b], errs[a])` (note: `b` first) puts bigger counts first.
Store it with `if c := ...; c != 0 { return c }`, then fall back to comparing
`a` and `b`. Add `"cmp"` to the imports.

### Answer

```go
slices.SortFunc(names, func(a, b string) int {
	if c := cmp.Compare(errs[b], errs[a]); c != 0 {
		return c
	}
	return cmp.Compare(a, b)
})
```

The counts decide when they differ; equal counts fall through to alphabetical order.
