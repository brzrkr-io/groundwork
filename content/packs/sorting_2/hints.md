## Task 1: Stop reordering the caller's slice

`slices.Sort(nodes)` sorts in place, and `nodes` is the caller's slice. Sort
something that belongs to the function instead.

### Closer

Make a copy with `slices.Clone(nodes)`, sort the copy, and return the copy
instead of `nodes`.

### Answer

```go
out := slices.Clone(nodes)
slices.Sort(out)
return out
```

The clone has its own storage, so sorting it leaves `nodes` alone.

## Task 2: Put the fullest disks first

`cmp.Compare(usage[a], usage[b])` puts small numbers first. One small change in
that call reverses the order. Leave the alphabetical tie-break as it is.

### Closer

Swapping the arguments of `cmp.Compare` reverses the order: compare
`usage[b]` with `usage[a]`.

### Answer

```go
if c := cmp.Compare(usage[b], usage[a]); c != 0 {
	return c
}
return cmp.Compare(a, b)
```

With `b` first, a bigger usage counts as "comes first".

## Task 3: Make ties predictable

When two nodes have the same pod count, the comparison returns `0`, and
`SortFunc` may leave them in any order. Add a second comparison for that case.

### Closer

Keep the pod comparison but only return it when it isn't `0`:
`if c := cmp.Compare(pods[a], pods[b]); c != 0 { return c }`. After that,
compare the names.

### Answer

```go
slices.SortFunc(out, func(a, b string) int {
	if c := cmp.Compare(pods[a], pods[b]); c != 0 {
		return c
	}
	return cmp.Compare(a, b)
})
```

Equal pod counts fall through to the name comparison, so ties come out alphabetical.
