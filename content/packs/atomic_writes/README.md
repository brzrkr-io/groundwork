# Atomic writes: temp file and rename

When your tool rewrites a config that nginx, systemd or the kubelet reads, a
half-written file can take a service down. Here you write files *atomically*:
readers see either the old file or the new one, never a mix.

**New here:** `os.CreateTemp`, `os.Rename`, `Sync`, `Chmod`, `os.Stat`, a deferred cleanup function.

## Learn

### What goes wrong with a plain write

`os.WriteFile` empties the target, then writes the new bytes. In between, the
file is empty or half-written. A crash, power loss or full disk leaves it
broken, and a service that reloads at the wrong moment reads garbage.

### The recipe: temp file, then rename

Renaming a file over another is *atomic* on Linux and macOS: anyone opening
the path gets the complete old file or the complete new one.

```text
  app.conf               (old content, still served to readers)
  1. CreateTemp   ->  app.conf.tmp-123456   (same directory, mode 0o600)
  2. Write        ->  new bytes into the temp file
  3. Chmod        ->  give it the right permissions
  4. Sync         ->  push the bytes from memory to the disk
  5. Close
  6. Rename       ->  app.conf.tmp-123456 replaces app.conf in one step
```

| Call | What it does |
|---|---|
| `os.CreateTemp(dir, pattern)` | creates a new, unique file in `dir` with mode `0o600` |
| `tmp.Write(data)` | writes the bytes; returns a count and an error |
| `tmp.Chmod(mode)` | sets the permission bits exactly (no umask) |
| `tmp.Sync()` | waits until the data is really on disk |
| `tmp.Close()` | closes the file; can report a write error |
| `os.Rename(old, new)` | moves `old` to `new`, replacing `new` if it is a file |
| `os.Remove(name)` | deletes a file |

### Naming the temp file

`filepath.Dir` and `filepath.Base` split a path. In the pattern, `*` is
replaced by a random string, so names never clash. `tmp.Name()` is the
temp file's full path.

```go
dir := filepath.Dir("/var/backups/db.sql")   // "/var/backups"
base := filepath.Base("/var/backups/db.sql") // "db.sql"
f, err := os.CreateTemp(dir, "upload-*.part") // e.g. /var/backups/upload-8812.part
```

The temp file **must be in the same directory** as the target. Rename is only
atomic within one filesystem, and `/tmp` is often a different one.

> **Why is it written like this?** Go requires the `*` to be inside the
> pattern string (it's text, so it goes in quotes). Starting the pattern with
> the target's name plus a suffix like `.tmp-*` is the author's choice: it
> makes leftovers easy to spot.

### Clean up on failure with a deferred function

If a step fails after the temp file exists, you must delete it, or failed runs
fill the directory with junk. Instead of cleaning up before every `return`,
`defer` one *anonymous function* and a flag that says "it worked":

```go
done := false
defer func() {
	if !done {
		fmt.Println("rolling back")
	}
}()
// ... steps that may return early ...
done = true
return nil
```

> **Why is it written like this?** `defer func() { ... }()` is Go syntax: you
> define a function and the final `()` calls it when the outer function
> returns. Closing an already-closed file is harmless, so the cleanup may
> close and remove without checking errors.

### Permissions and os.Stat

`CreateTemp` always makes `0o600` files, and the rename keeps the temp file's
mode. `os.Stat(path)` returns info about an existing file (and an error if it
is missing); `info.Mode().Perm()` gives its permission bits.

### How the tests check "replaced, not rewritten"

`os.SameFile(a, b)` reports whether two `os.Stat` results are the same file on
disk (on Linux, the same *inode*). Writing in place keeps the same file; a
rename swaps in a new one. The tests require the file after your write to be
a **different** file from before.

## Your task

Your config tool needs `WriteFileAtomic(path, data, perm)`: same signature as
`os.WriteFile`, but safe. The starter only imports `io/fs`; add `os` and
`path/filepath`.

### 1. Write through a temp file and rename

**Goal:** `WriteFileAtomic` writes `data` to a temp file next to `path`, then renames it over `path`.

- Create the temp file in `filepath.Dir(path)`.
- Then: write `data`, `Sync`, `Close`, `os.Rename` to `path`.
- Check every step's error and return the first one.
- On success, `path` holds exactly `data` (empty `data` gives an empty
  file), an existing file is replaced (not rewritten in place), and no temp
  file is left.

```go
WriteFileAtomic(dir+"/app.conf", []byte("port = 8080\n"), 0o644)
// dir holds only app.conf, containing "port = 8080\n"
```

**Think it through:**
- Follow steps 1, 2, 4, 5, 6 of the recipe (step 3 is task 3).
- Which method gives you the temp file's path for `os.Rename`?

### 2. Clean up when something fails

**Goal:** if any step fails, return a non-nil error and remove the temp file.

The directory must hold exactly what it held before. The tests use two
failures:

- `path` is an existing, non-empty directory, so the rename fails;
- the parent directory doesn't exist, so `CreateTemp` fails.

**Think it through:**
- From which line on does a temp file exist that could be left behind?
- Where does the "it worked" flag get set to true: before or after the rename?

### 3. Set the permissions

**Goal:** before the rename, `Chmod` the temp file to the right mode.

- If `path` already exists, keep its current permission bits, whatever
  `perm` says.
- If it doesn't exist, use `perm` exactly.

```go
// secret.env exists with mode 0o600
WriteFileAtomic(dir+"/secret.env", data, 0o644) // still 0o600
// new.conf doesn't exist
WriteFileAtomic(dir+"/new.conf", data, 0o640)   // now 0o640
```

**Think it through:**
- Look at the target *before* you replace it: which call tells you if it
  exists and its mode?
- Start from the `perm` parameter and change it only when the file exists.

## Learn more

- [os.CreateTemp](https://pkg.go.dev/os#CreateTemp)
- [os.Rename](https://pkg.go.dev/os#Rename)
- [os.File.Sync](https://pkg.go.dev/os#File.Sync)
- [os.SameFile](https://pkg.go.dev/os#SameFile)
