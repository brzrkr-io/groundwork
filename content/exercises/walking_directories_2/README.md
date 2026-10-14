# Walking directories: fix the backup cleaner

A teammate started a tool that tidies the backup volume on a database host.
It walks the real disk with `filepath.WalkDir`, and it has two bugs and one
missing function.

**New here:** `filepath.WalkDir`, `filepath.Rel`, adding up values in a map.

## Learn

### Recap

- `fs.WalkDir(fsys, ".", fn)` calls `fn(p, d, err)` once per entry.
- Check `err` first: `d` can be `nil` when reading fails.
- Return `nil` to keep going, `fs.SkipDir` to stay out of a directory, any
  other error to stop.
- `d.Info()` gives an `fs.FileInfo` with `Size()` and `ModTime()`.

### filepath.WalkDir: walking the real disk

`path/filepath` has a twin of `fs.WalkDir` that takes a directory path instead
of an `fs.FS`. The callback has exactly the same shape:

```go
err := filepath.WalkDir("/srv/backups", func(p string, d fs.DirEntry, err error) error {
	// p is "/srv/backups", then "/srv/backups/db/mon.sql.gz", ...
	return nil
})
```

Paths start with the root you passed in and use the OS separator (`\` on
Windows). If the root doesn't exist, the very first callback gets
`p == root`, `d == nil` and a non-nil `err`.

### Relative paths

`filepath.Rel(base, target)` cuts `base` off the front of `target`.
`filepath.ToSlash` then turns `\` into `/`, so results read the same
everywhere:

```go
rel, err := filepath.Rel("/home", "/home/ana/notes.txt")
// rel == "ana/notes.txt" (on Windows: `ana\notes.txt`)
rel = filepath.ToSlash(rel) // always "ana/notes.txt"
```

### Splitting at the first separator: strings.Cut

`strings.Cut(s, sep)` splits `s` at the **first** `sep`. It returns the part
before, the part after, and `found` (false when `sep` isn't in `s`; then
`before` is all of `s`):

```go
user, host, found := strings.Cut("ana@web01", "@") // "ana", "web01", true
user, host, found = strings.Cut("ana", "@")        // "ana", "", false
```

> **Why is it written like this?** Go requires one variable per returned
> value; use `_` for one you don't need. The names are your choice.

### Adding up into a map

A map must be created with `make` before you write to it. A missing key reads
as `0`, so `+=` works on the first visit too:

```go
counts := make(map[string]int)
counts["web"] += 3 // was 0, now 3
counts["web"] += 2 // now 5
```

> **Why is it written like this?** Go requires `make` (or a `{}` literal):
> writing to a `nil` map crashes. Reading a missing key giving zero is a Go
> rule, which is why no "does it exist yet?" check is needed.

### Modification times and `now`

`info.ModTime()` is when the file was last written. `now.Sub(t)` gives the
age as a `time.Duration`. The functions take `now` as a parameter instead of
calling `time.Now()`, so tests can pick a fixed time and set file times with
`os.Chtimes`.

### Tests on the real disk

`t.TempDir()` gives each test a fresh directory that is deleted afterwards.
The tests build small trees in it with `os.MkdirAll` and `os.WriteFile`.

## Your task

Backups live under a root directory, one folder per database. Folders named
`pinned` hold backups that must never be deleted.

`StaleFiles(root, now, maxAge)` returns the files under `root` that are
**more than** `maxAge` old (`now.Sub(modTime) > maxAge`), as sorted
forward-slash paths relative to `root`. A file exactly `maxAge` old is not
stale yet. That logic already works; tasks 1 and 2 fix the rest.

### 1. Don't crash on a missing root

**Goal:** when `root` doesn't exist, `StaleFiles` returns an error instead of panicking.

- Return the walk's own error (wrapping it with `%w` is fine).
- The tests check it with `errors.Is(err, fs.ErrNotExist)`.

```go
StaleFiles("/no/such/dir", now, 24*time.Hour)
// => nil, error (no panic)
```

**Think it through:**
- Which callback parameter holds the problem? Where is it checked right now?
- Which line crashes when `d` is `nil`? Your check must come before it.

### 2. Really skip pinned folders

**Goal:** files inside any directory named `pinned`, at any depth, never appear in `StaleFiles`' results.

```go
// root/db/old.sql.gz         (10 days old)
// root/db/pinned/gold.sql.gz (90 days old)
StaleFiles(root, now, 7*24*time.Hour)
// => ["db/old.sql.gz"]
```

**Think it through:**
- The code already spots `pinned`. What does its `return` value tell the
  walk to do?
- Which return value from the table in the recap means "don't go inside"?

### 3. Total the size of each folder

**Goal:** `DirSizes(root)` returns a map from each top-level folder under `root` to the total bytes of all files inside it, at any depth.

- Files directly in `root` are added up under the key `"."`.
- Folders containing no files don't appear in the map.
- `pinned` folders are counted like any other.
- Add `"strings"` to the imports.

```go
// root/notes.txt             (10 bytes)
// root/db/mon.sql.gz         (300 bytes)
// root/db/pinned/gold.sql.gz (200 bytes)
// root/redis/dump.rdb        (50 bytes)
DirSizes(root)
// => map[".":10 "db":500 "redis":50]
```

**Think it through:**
- For `db/pinned/gold.sql.gz`, the key is `db`: the part before the first
  `/` of the relative path. What if there is no `/`?
- What should the callback do when `d` is a directory? Only files have sizes
  worth adding.

## Learn more

- [Package path/filepath: WalkDir](https://pkg.go.dev/path/filepath#WalkDir)
- [Package path/filepath: Rel](https://pkg.go.dev/path/filepath#Rel)
- [Package strings: Cut](https://pkg.go.dev/strings#Cut)
- [Package os: Chtimes](https://pkg.go.dev/os#Chtimes)
