## Task 1: Fix Sum

Look at the line inside the loop: it replaces `total` each time instead of adding to it.

### Closer

Use the `+=` operator, which adds the right side onto the variable. It works because `Number` only allows number types.

### Answer

```go
total += v
```

`total` starts at the zero value of `T`, so an empty slice sums to `0`.

## Task 2: List map keys in order

Collect every key of the map into `keys` first, then sort the whole slice once at the end.

### Closer

`for k := range m` gives you each key; `append` it to `keys`. After the loop, call `slices.Sort(keys)`, and add `"slices"` to the imports.

### Answer

```go
for k := range m {
	keys = append(keys, k)
}
slices.Sort(keys)
```

`slices.Sort` works because `K` is `cmp.Ordered`.

## Task 3: Fix GroupBy

`groups[k] = []T{item}` replaces the whole group with a new one-item slice. You want to add the item to the group that's already there.

### Closer

Use `append` on the existing group: `append(groups[k], item)`. A missing key reads as a `nil` slice, and appending to `nil` works fine.

### Answer

```go
groups[k] = append(groups[k], item)
```

Store the result back, because `append` may return a new slice.
