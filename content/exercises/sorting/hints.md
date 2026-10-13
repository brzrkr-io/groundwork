## Task 1: copy, then sort

`slices.Sort` rearranges the slice you give it. Make a copy with
`slices.Clone(hosts)`, sort the copy, and return the copy.

## Task 2: one line

`maps.Keys(errs)` gives you the keys, and `slices.Sorted(...)` turns them into
a sorted slice. Don't forget to add `"maps"` to the `import` block.

## Task 3: reverse, then tie-break

`cmp.Compare(errs[b], errs[a])` (note: `b` first) puts bigger counts first.
Only when that returns `0` do you need to look at the names.

## Task 3: almost the answer

```go
slices.SortFunc(names, func(a, b string) int {
	if c := cmp.Compare(errs[b], errs[a]); c != 0 {
		return c
	}
	return cmp.Compare(a, b)
})
```
