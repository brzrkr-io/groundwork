# Walking directories: find big log files

Disks fill up, and the culprit is usually a forgotten log file somewhere deep
in a directory tree. In this exercise you visit every file under a directory,
the way `find` and `du` do, and pick out the big log files.

## Learn

### Walking a tree with a callback

The `io/fs` package has a function that visits every file and directory
under a starting point, one at a time:

```go
err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
	if err != nil {
		return err // something went wrong reading p: stop the walk
	}
	fmt.Println(p, d.IsDir())
	return nil // carry on with the next entry
})
```

`fs.WalkDir` calls your function (the *callback*) once for every entry it
finds, starting with the root `"."` itself. The callback gets three values:

| Parameter | What it is |
|---|---|
| `p string` | The entry's path, such as `"logs/app.log"`. |
| `d fs.DirEntry` | Facts about the entry: `d.Name()`, `d.IsDir()`, `d.Info()`. |
| `err error` | Non-nil if the walk couldn't read this entry. |

**Always check `err` first.** When reading fails, `d` may be `nil`, and
calling `d.IsDir()` on a nil value crashes your program.

`d.Name()` is only the last part of the path (`"app.log"`). To get the file
size, call `d.Info()`. It returns an `fs.FileInfo` and an error:

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
| `nil` | Keep walking. |
| `fs.SkipDir` | When `d` is a directory: don't go inside it. The walk continues elsewhere. |
| any other error | Stop the walk. `fs.WalkDir` returns that error. |

`fs.SkipDir` is how you avoid directories you don't care about, such as
`.git` or `node_modules`:

```go
if d.IsDir() && d.Name() == ".git" {
	return fs.SkipDir
}
```

### fs.FS: a filesystem you can swap out

`fs.WalkDir` doesn't walk your disk directly. Its first argument is an
`fs.FS`, an interface for "something that holds files". You choose which
one:

```go
fsys := os.DirFS("/var/log") // the real directory, in main
```

```go
fsys := fstest.MapFS{ // an in-memory fake, in tests
	"app.log":       {Data: []byte("hello")},
	"nginx/err.log": {Data: make([]byte, 2048)}, // 2048 zero bytes
}
```

This is dependency injection, as in *Fakes and dependency injection*: your
function takes the interface, so tests never touch the real disk.

Paths inside an `fs.FS` always use forward slashes (`"nginx/err.log"`), even
on Windows. So use the `path` package (`path.Match`, `path.Ext`) with them,
not `path/filepath`.

`path.Match(pattern, name)` checks a shell-style pattern, where `*` matches any
run of characters:

```go
ok, _ := path.Match("*.log", "app.log")   // ok == true
ok, _ = path.Match("*.log", "app.log.1")  // ok == false
```

It only returns an error for a broken pattern, so with a fixed pattern like
`"*.log"` you can ignore it with `_`.

### Walk order isn't sorted order

`fs.WalkDir` reads each directory in name order. But the full paths it
produces are not sorted overall. A directory `a` comes before the file
`a.txt`, so you get `"a/b.log"` before `"a.txt"`. As plain strings, `"a.txt"`
sorts first, because `.` comes before `/`. When you want sorted output, collect
the paths and call `slices.Sort` at the end.

## Your task

Your team's build servers keep running out of disk. Write the helpers for a
cleanup tool that lists files and finds oversized logs.

### 1. List every file

Complete `ListFiles`. It returns the path of every file in `fsys` (skip
directories), sorted with `slices.Sort`. Add `"slices"` to the imports.

```go
ListFiles(fstest.MapFS{
	"a.txt":   {},
	"a/b.log": {},
})
// => ["a.txt" "a/b.log"]
```

An empty filesystem gives an empty (or nil) slice.

### 2. Skip .git directories

`.git` directories hold thousands of files nobody wants in a cleanup report.
Make `ListFiles` skip every directory named `.git`, at any depth, by
returning `fs.SkipDir`.

```go
ListFiles(fstest.MapFS{
	"README.md":       {},
	".git/HEAD":       {},
	"svc/.git/config": {},
	"svc/main.go":     {},
})
// => ["README.md" "svc/main.go"]
```

### 3. Find large log files

Complete `LargeLogs`. It returns the sorted paths of files whose name matches
`*.log` (use `path.Match` on `d.Name()`) and whose size is **at least**
`minSize` bytes. Like `ListFiles`, it skips `.git` directories. Add `"path"`
to the imports.

```go
LargeLogs(fstest.MapFS{
	"api/app.log":   {Data: make([]byte, 500)},
	"api/app.log.1": {Data: make([]byte, 900)},
	"db/slow.log":   {Data: make([]byte, 100)},
	"web/big.txt":   {Data: make([]byte, 900)},
}, 100)
// => ["api/app.log" "db/slow.log"]
```

`app.log.1` doesn't match `*.log`, and `big.txt` isn't a log. A file of
exactly `minSize` bytes counts.

## Learn more

- [Package io/fs: WalkDir](https://pkg.go.dev/io/fs#WalkDir)
- [Package testing/fstest: MapFS](https://pkg.go.dev/testing/fstest#MapFS)
- [Package path: Match](https://pkg.go.dev/path#Match)
- [Package os: DirFS](https://pkg.go.dev/os#DirFS)
