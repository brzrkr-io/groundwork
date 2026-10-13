## Task 1: collect, then sort

Inside the callback, `d.IsDir()` tells you whether `p` is a directory. Only
append when it's **not** one. After `fs.WalkDir` returns, sort the whole
slice once with `slices.Sort(files)`.

## Task 2: skip by name

`d.Name()` is just the last part of the path, so `svc/.git` has the name
`".git"`. Check it **before** the append, and return `fs.SkipDir` instead of
`nil`:

```go
if d.IsDir() && d.Name() == ".git" {
	return fs.SkipDir
}
```

## Task 3: the shape of the callback

Copy the `fs.WalkDir` call from `ListFiles`, with its own `logs` slice. In
the callback: handle `err`, skip `.git`, return `nil` for other directories,
then check the name with `path.Match("*.log", d.Name())`, then the size.

## Task 3: almost the answer

```go
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
```

Don't forget `slices.Sort(logs)` after the walk.
