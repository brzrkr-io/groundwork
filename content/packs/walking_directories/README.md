# Walking directories: find big log files

Disks fill up, and the culprit is usually a forgotten log file deep in a
directory tree. Here you visit every file under a directory, the way `find`
and `du` do, and pick out the big log files.

**New here:** `fs.WalkDir`, callbacks, `fs.SkipDir`, `fs.FS`, `path.Match`.

## Learn

### Walking a tree with a callback

`fs.WalkDir` visits every file and directory under a starting point, one at a
time. For each one it calls a function you give it, the *callback*:

```go
err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
	if err != nil {
		return err // couldn't read p: stop the walk
	}
	fmt.Println(p, d.IsDir())
	return nil // carry on with the next entry
})
```

The first call is for the root `"."` itself. The callback gets:

| Parameter | What it is |
|---|---|
| `p string` | the entry's path, such as `"logs/app.log"` |
| `d fs.DirEntry` | facts about it: `d.Name()`, `d.IsDir()`, `d.Info()` |
| `err error` | non-nil if the walk couldn't read this entry |

> **Why is it written like this?** Go requires the callback to have exactly
> this shape: `(string, fs.DirEntry, error)` in, `error` out. The parameter
> names (`p`, `d`) are the author's choice. Writing the function right inside
> the call is convention for short callbacks.

**Always check `err` first.** When reading fails, `d` may be `nil`, and
`d.IsDir()` on `nil` crashes your program.

### Name and size

`d.Name()` is only the last part of the path (`"app.log"`). For the size,
`d.Info()` returns an `fs.FileInfo` and an error:

```go
info, err := d.Info()
if err != nil {
	return err
}
size := info.Size() // bytes, as an int64
```

### What the callback returns

| Return value | Effect |
|---|---|
| `nil` | keep walking |
| `fs.SkipDir` | when `d` is a directory: don't go inside it; keep walking elsewhere |
| any other error | stop; `fs.WalkDir` returns that error |

```go
if d.IsDir() && d.Name() == "node_modules" {
	return fs.SkipDir
}
```

### fs.FS: a filesystem you can swap out

`fs.WalkDir`'s first argument is an `fs.FS`, an interface for "something that
holds files". `main` passes the real disk; tests pass a fake in memory:

```go
fsys := os.DirFS("/var/log") // the real directory

fsys := fstest.MapFS{ // an in-memory fake
	"app.log":       {Data: []byte("hello")},
	"nginx/err.log": {Data: make([]byte, 2048)}, // 2048 zero bytes
}
```

This is dependency injection, as in *Fakes and dependency injection*.

### Paths and patterns

Paths in an `fs.FS` always use forward slashes, even on Windows. So use the
`path` package with them, not `path/filepath`.

`path.Match(pattern, name)` checks a shell-style pattern; `*` matches any run
of characters:

```go
ok, _ := path.Match("*.conf", "nginx.conf")   // true
ok, _ = path.Match("*.conf", "nginx.conf.bak") // false
```

> **Why is it written like this?** `path.Match` returns `(bool, error)` and Go
> requires you to accept both. The error only happens for a broken pattern,
> so with a fixed pattern ignoring it with `_` is fine (author's choice).

### Walk order isn't sorted order

`fs.WalkDir` reads each directory in name order, but full paths come out
unsorted overall: directory `a` (and `a/b.log`) comes before file `a.txt`, yet
as strings `"a.txt"` sorts first, because `.` comes before `/`. For sorted
output, collect the paths and call `slices.Sort` once at the end.

## Your task

Your build servers keep running out of disk. Write the helpers for a cleanup
tool that lists files and finds oversized logs.

### 1. List every file

**Goal:** `ListFiles(fsys)` returns the path of every file (not directory) in `fsys`, sorted.

- Sort with `slices.Sort`; add `"slices"` to the imports.
- An empty filesystem gives an empty (or `nil`) slice.

```go
ListFiles(fstest.MapFS{
	"a.txt":   {},
	"a/b.log": {},
})
// => ["a.txt" "a/b.log"]
```

**Think it through:**
- Which method of `d` tells you it's a directory?
- Sort inside the callback, or once after the walk?

### 2. Skip .git directories

**Goal:** `ListFiles` skips every directory named `.git`, at any depth, by returning `fs.SkipDir`.

A *file* named `.gitignore` is still listed.

```go
ListFiles(fstest.MapFS{
	"README.md":       {},
	".git/HEAD":       {},
	"svc/.git/config": {},
	"svc/main.go":     {},
})
// => ["README.md" "svc/main.go"]
```

**Think it through:**
- Do you compare the whole path `p`, or just the name? Which works at any depth?
- Two things must both be true. Which ones?

### 3. Find large log files

**Goal:** `LargeLogs(fsys, minSize)` returns the sorted paths of `*.log` files of at least `minSize` bytes.

- The file's name (`d.Name()`) must match `*.log` with `path.Match`.
- Its size must be **at least** `minSize`: exactly `minSize` counts.
- Skip `.git` directories, like `ListFiles`.
- No matches gives an empty (or `nil`) slice.
- Add `"path"` to the imports.

```go
LargeLogs(fstest.MapFS{
	"api/app.log":   {Data: make([]byte, 500)},
	"api/app.log.1": {Data: make([]byte, 900)},
	"db/slow.log":   {Data: make([]byte, 100)},
	"web/big.txt":   {Data: make([]byte, 900)},
}, 100)
// => ["api/app.log" "db/slow.log"]
```

`app.log.1` doesn't match `*.log`, and `big.txt` isn't a log.

**Think it through:**
- Start from the shape of `ListFiles`: what changes, what stays?
- Check the name before calling `d.Info()`. Which comparison operator means
  "at least"?

## Learn more

- [Package io/fs: WalkDir](https://pkg.go.dev/io/fs#WalkDir)
- [Package testing/fstest: MapFS](https://pkg.go.dev/testing/fstest#MapFS)
- [Package path: Match](https://pkg.go.dev/path#Match)
- [Package os: DirFS](https://pkg.go.dev/os#DirFS)
