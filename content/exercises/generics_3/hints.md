## Task 1: the last element

The newest item is the last one in `r.items`, at index `len(r.items)-1`.
Check for an empty ring first and return `var zero T` with `false`.

## Task 2: trim after appending

After the `append`, compare `len(r.items)` with `r.limit`. If there are too
many, re-slice so the first item is left out: `r.items[1:]`.

## Task 3: clone the slice

`slices.Clone(r.items)` returns a new slice with the same elements and its own
backing array. Add `"slices"` to the imports.

## Almost the answer

```go
if len(r.items) == 0 {
	var zero T
	return zero, false
}
return r.items[len(r.items)-1], true
```

```go
r.items = append(r.items, v)
if len(r.items) > r.limit {
	r.items = r.items[1:]
}
```

```go
return slices.Clone(r.items)
```
