# Walking directories: fix the backup cleaner

A teammate started a tool that tidies the backup volume on a database host.
It walks the real disk with `filepath.WalkDir`, and it has two bugs and one
missing function.

## Learn

### Recap

`fs.WalkDir(fsys, ".", fn)` calls `fn(p, d, err)` once for every entry under
the root. Check `err` first, because `d` can be `nil` when reading fails.
Return `nil` to keep going, `fs.SkipDir` to stay out of a directory, or any
other error to stop the walk. `d.Info()` gives you an `fs.FileInfo` with the
size.

### filepath.WalkDir: walking the real disk

`path/filepath` has a twin of `fs.WalkDir` that takes a directory path
instead of an `fs.FS`. The callback is exactly the same:

```go
err := filepath.WalkDir("/srv/backups", func(p string, d fs.DirEntry, err error) error {
	// p is "/srv/backups", then "/srv/backups/db/mon.sql.gz", ...
	return nil
})
```

The paths start with the root you passed in and use the operating system's
separator (`\` on Windows). To report a path relative to the root, use
`filepath.Rel`, then `filepath.ToSlash` so it reads the same on every system:

```go
rel, err := filepath.Rel("/srv/backups", "/srv/backups/db/mon.sql.gz")
// rel == "db/mon.sql.gz" (on Windows: `db\mon.sql.gz`)
rel = filepath.ToSlash(rel) // always "db/mon.sql.gz"
```

If the root itself doesn't exist, the very first callback gets `p == root`,
`d == nil` and a non-nil `err`.

### Modification times

`info.ModTime()` is the `time.Time` the file was last written. To ask "how
old is it?", subtract it from the current time:

```go
age := now.Sub(info.ModTime()) // a time.Duration
```

The functions here take `now` as a parameter instead of calling
`time.Now()`. That way tests can pick a fixed "now" and set file times with
`os.Chtimes(name, accessTime, modTime)`, and the results never depend on the
clock.

### Tests on the real disk

`t.TempDir()` gives each test a fresh, empty directory that is deleted when
the test ends. The tests in this exercise build small trees in it with
`os.MkdirAll` and `os.WriteFile`.

## Your task

Backups live under a root directory, one folder per database. Folders named
`pinned` hold backups that must never be deleted.

### 1. Don't crash on a missing root

`StaleFiles(root, now, maxAge)` returns the files under `root` that are
**more than** `maxAge` old (`now.Sub(modTime) > maxAge`), as sorted
forward-slash paths relative to `root`. When `root` doesn't exist, it
panics with a nil pointer error. Fix the callback so it returns the walk's
error instead (wrapping it with `%w` is fine). The tests check it with
`errors.Is(err, fs.ErrNotExist)`.

```go
StaleFiles("/no/such/dir", now, 24*time.Hour)
// => nil, error (no panic)
```

A file exactly `maxAge` old is not stale yet.

### 2. Really skip pinned folders

`StaleFiles` is meant to leave directories named `pinned` out of the
results, but old files inside them still show up. Fix it.

```go
// root/db/old.sql.gz         (10 days old)
// root/db/pinned/gold.sql.gz (90 days old)
StaleFiles(root, now, 7*24*time.Hour)
// => ["db/old.sql.gz"]
```

### 3. Total the size of each folder

Write `DirSizes(root)`, a small `du`. It returns a map from each top-level
folder under `root` to the total size in bytes of all the files inside it, at
any depth. Files directly in `root` are added up under the key `"."`.
Folders that contain no files don't appear in the map. `pinned` folders are
counted like any other. You'll need `filepath.Rel`, `filepath.ToSlash` and
`strings.Cut` (add `"strings"` to the imports).

```go
// root/notes.txt             (10 bytes)
// root/db/mon.sql.gz         (300 bytes)
// root/db/pinned/gold.sql.gz (200 bytes)
// root/redis/dump.rdb        (50 bytes)
DirSizes(root)
// => map[".":10 "db":500 "redis":50]
```

## Learn more

- [Package path/filepath: WalkDir](https://pkg.go.dev/path/filepath#WalkDir)
- [Package path/filepath: Rel](https://pkg.go.dev/path/filepath#Rel)
- [Package os: Chtimes](https://pkg.go.dev/os#Chtimes)
