## Task 1: Ignore AWS-managed tags

`ChangedKeys` has two loops, one over `want` and one over `got`. Before you
compare a key in either loop, decide whether it belongs to AWS and should be
skipped.

### Closer

At the top of each loop body, test the key with `strings.HasPrefix(k, "aws:")`
and use `continue` to jump to the next key. Add `"strings"` to the imports.

### Answer

```go
for k, v := range want {
	if strings.HasPrefix(k, "aws:") {
		continue
	}
	if gv, ok := got[k]; !ok || gv != v {
		keys = append(keys, k)
	}
}
for k := range got {
	if strings.HasPrefix(k, "aws:") {
		continue
	}
	if _, ok := want[k]; !ok {
		keys = append(keys, k)
	}
}
```

Skipping on both sides means an `aws:` key is never counted, whichever map has it.

## Task 2: Make the report order stable

Map order is random, so `changes` comes out in a different order each run.
Sort the slice once, just before `Diff` returns it.

### Closer

Use `slices.SortFunc` with a compare function that compares the `ID` fields
with `cmp.Compare`. Add the `"cmp"` import.

### Answer

```go
slices.SortFunc(changes, func(a, b Change) int {
	return cmp.Compare(a.ID, b.ID)
})
return changes
```

Sorting by `ID` gives the same report every night, whatever order the maps used.
