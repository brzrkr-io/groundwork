## Task 1: add, don't replace

`total = v` throws away what `total` held before. You want to add `v` on top
of it.

## Task 2: collect, then sort

Range over the map to get each key (`for k := range m`) and `append` it to
`keys`. After the loop, call `slices.Sort(keys)`. It works because `K` is
`cmp.Ordered`. Remember to add `"slices"` to the imports.

## Task 3: grow the group

`groups[k] = []T{item}` replaces the whole group with a new one-item slice.
Use `append` on the existing group instead. A missing key reads as a `nil`
slice, and appending to `nil` works fine.

## Almost the answer

```go
total += v
```

```go
for k := range m {
	keys = append(keys, k)
}
slices.Sort(keys)
```

```go
groups[k] = append(groups[k], item)
```
