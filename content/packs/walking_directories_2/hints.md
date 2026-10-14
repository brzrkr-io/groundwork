## Task 1: Don't crash on a missing root

When the root is missing, the callback is called with a non-nil `err` and `d == nil`. The callback uses `d` before it ever looks at `err`.

### Closer

Put the usual error check at the very top of the callback, before `d.IsDir()`, and return the error so `filepath.WalkDir` passes it back.

### Answer

```go
if err != nil {
	return err
}
```

## Task 2: Really skip pinned folders

The callback spots `pinned`, but returning `nil` for a directory means "fine, keep going", so the walk still goes inside it.

### Closer

The return value that means "don't go into this directory" is `fs.SkipDir`. Use it in the `d.Name() == "pinned"` branch.

### Answer

```go
if d.IsDir() {
	if d.Name() == "pinned" {
		return fs.SkipDir
	}
	return nil
}
```

## Task 3: Total the size of each folder

Walk `root` with a callback like `StaleFiles`'s, but add each file's size to a map. The key is the first part of the file's path relative to `root`, or `"."` when the file sits directly in `root`.

### Closer

Create the map with `make(map[string]int64)` before the walk. For each file, `filepath.Rel(root, p)` then `strings.Cut(filepath.ToSlash(rel), "/")` gives the top-level folder; `found` is false when there is no `/`. Add `info.Size()` with `sizes[top] +=`.

### Answer

```go
sizes := make(map[string]int64)
err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		return nil
	}
	info, err := d.Info()
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return err
	}
	top, _, found := strings.Cut(filepath.ToSlash(rel), "/")
	if !found {
		top = "."
	}
	sizes[top] += info.Size()
	return nil
})
if err != nil {
	return nil, err
}
return sizes, nil
```
