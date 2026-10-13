## Task 1: err comes first

When the root is missing, the callback is called with `d == nil` and a
non-nil `err`. Put the usual check at the very top of the callback:

```go
if err != nil {
	return err
}
```

## Task 2: nil versus SkipDir

Returning `nil` for a directory means "fine, keep going", so the walk still
goes inside it. The return value that means "don't go in here" is
`fs.SkipDir`.

## Task 3: find the top-level folder

Skip directories, then turn `p` into a path relative to `root`. The part
before the first `/` is the top-level folder. When there is no `/`, the file
sits directly in `root`:

```go
rel, err := filepath.Rel(root, p)
if err != nil {
	return err
}
top, _, found := strings.Cut(filepath.ToSlash(rel), "/")
if !found {
	top = "."
}
```

## Task 3: adding up

Create the map with `make(map[string]int64)` before the walk, so you can
write to it. Inside the callback, after finding `top`, add the size:

```go
info, err := d.Info()
if err != nil {
	return err
}
sizes[top] += info.Size()
```

Return the map after `filepath.WalkDir` finishes without an error.
