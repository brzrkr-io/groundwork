## Task 1: Return the newest item

The newest item is the last one in `r.items`. Check for an empty ring first, because there is no last item then.

### Closer

The last element is at index `len(r.items)-1`. For the empty case, `var zero T` then `return zero, false`.

### Answer

```go
if len(r.items) == 0 {
	var zero T
	return zero, false
}
return r.items[len(r.items)-1], true
```

## Task 2: Forget the oldest item

After the `append`, check whether the ring now holds too many items. If it does, drop the first one.

### Closer

Compare `len(r.items)` with `r.limit`. Re-slicing with `r.items[1:]` leaves out the first (oldest) item.

### Answer

```go
r.items = append(r.items, v)
if len(r.items) > r.limit {
	r.items = r.items[1:]
}
```

Since each push adds only one item, dropping one is always enough.

## Task 3: Hand out a copy

Returning `r.items` hands the caller the ring's own backing array, so their changes leak in. Return a new slice with the same elements instead.

### Closer

`slices.Clone(r.items)` makes a new slice with its own backing array. Add `"slices"` to the imports.

### Answer

```go
return slices.Clone(r.items)
```
