## Read each failure message

Each test names the function and shows what it returned. Compare that with the
README: is the order backwards, are ties out of order, or did the caller's
slice change?

## Task 1: who owns the slice?

`slices.Sort(nodes)` reorders the caller's slice. Sort a
`slices.Clone(nodes)` instead and return that.

## Task 2: swap the arguments

`cmp.Compare(usage[a], usage[b])` puts small numbers first. Swap `a` and `b`
in that one call. Leave the alphabetical tie-break as it is.

## Task 3: add the tie-break

```go
if c := cmp.Compare(pods[a], pods[b]); c != 0 {
	return c
}
return cmp.Compare(a, b)
```
