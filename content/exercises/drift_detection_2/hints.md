## Task 1: skip before you compare

`ChangedKeys` has two loops, one over `want` and one over `got`. At the top
of each loop body, skip the key when it is AWS-managed. `continue` jumps to
the next key.

## Task 1: the check

```go
for k := range got {
	if strings.HasPrefix(k, "aws:") {
		continue
	}
	if _, ok := want[k]; !ok {
		keys = append(keys, k)
	}
}
```

Do the same in the loop over `want`.

## Task 2: sort by ID

Add this just before `return changes`, and add the `cmp` import:

```go
slices.SortFunc(changes, func(a, b Change) int {
	return cmp.Compare(a.ID, b.ID)
})
```
