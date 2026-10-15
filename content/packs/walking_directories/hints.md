## Task 1: List every file

Inside the callback, `d.IsDir()` tells you whether `p` is a directory. Collect only the paths that aren't, then sort the whole slice once after the walk.

### Closer

Append with `if !d.IsDir() { files = append(files, p) }`. After `fs.WalkDir` returns without an error, call `slices.Sort(files)`. Add `"slices"` to the imports.

### Answer

```go
// inside the callback, after the err check:
if !d.IsDir() {
	files = append(files, p)
}
return nil

// after the walk:
slices.Sort(files)
return files, nil
```

## Task 2: Skip .git directories

`d.Name()` is just the last part of the path, so `svc/.git` has the name `".git"`. Check for it before you append, and return a value that tells the walk not to go inside.

### Closer

Returning `fs.SkipDir` from the callback for a directory skips everything in it. Test `d.IsDir() && d.Name() == ".git"`.

### Answer

```go
if d.IsDir() && d.Name() == ".git" {
	return fs.SkipDir
}
```

## Task 3: Find large log files

Copy the `fs.WalkDir` call from `ListFiles`, with its own `logs` slice. In the callback: handle `err`, skip `.git`, ignore other directories, then check the name, then the size.

### Closer

`path.Match("*.log", d.Name())` returns `(bool, error)`. For the size, `d.Info()` gives an `fs.FileInfo` with `.Size()`; keep the file when it is `>= minSize`. Sort with `slices.Sort(logs)` and add `"path"` to the imports.

### Answer

```go
var logs []string
err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		if d.Name() == ".git" {
			return fs.SkipDir
		}
		return nil
	}
	if ok, _ := path.Match("*.log", d.Name()); !ok {
		return nil
	}
	info, err := d.Info()
	if err != nil {
		return err
	}
	if info.Size() >= minSize {
		logs = append(logs, p)
	}
	return nil
})
if err != nil {
	return nil, err
}
slices.Sort(logs)
return logs, nil
```
