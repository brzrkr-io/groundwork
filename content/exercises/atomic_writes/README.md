# Atomic writes: temp file and rename

When your tool rewrites a config that nginx, systemd or the kubelet reads, a
half-written file can take a service down. In this exercise you write files
*atomically*: readers see either the old file or the new one, never a mix.

## Learn

### What goes wrong with a plain write

`os.WriteFile` truncates the target and then writes the new bytes into it. In
between, the file is empty or half-written. If the process crashes, the
machine loses power, or the disk fills up halfway, the broken file stays. Even
without a crash, a service that reloads at the wrong moment reads partial
content.

### The recipe: write a temp file, then rename

Renaming a file over another one is *atomic* on Linux and macOS: anyone who
opens the path gets either the complete old file or the complete new one.

```text
  app.conf               (old content, still served to readers)
  1. CreateTemp   ->  app.conf.tmp-123456   (same directory, mode 0o600)
  2. Write        ->  new bytes into the temp file
  3. Chmod        ->  give it the right permissions
  4. Sync         ->  push the bytes from memory to the disk
  5. Close
  6. Rename       ->  app.conf.tmp-123456 replaces app.conf in one step
```

```go
dir := filepath.Dir(path)   // "/etc/myapp" for "/etc/myapp/app.conf"
base := filepath.Base(path) // "app.conf"

tmp, err := os.CreateTemp(dir, base+".tmp-*")
// tmp is an *os.File; tmp.Name() is its full path.
// The "*" is replaced by a random string, so names never clash.
```

| Call | What it does |
|---|---|
| `os.CreateTemp(dir, pattern)` | creates a new, unique file in `dir` with mode `0o600` |
| `tmp.Write(data)` | writes the bytes |
| `tmp.Chmod(mode)` | sets the permission bits exactly (the umask does not apply) |
| `tmp.Sync()` | waits until the data is really on disk, not just in memory |
| `tmp.Close()` | closes the file and can report a write error |
| `os.Rename(old, new)` | moves `old` to `new`, replacing `new` if it is a file |
| `os.Remove(name)` | deletes a file |

The temp file **must be in the same directory** as the target. A rename is
only atomic within one filesystem, and `/tmp` is often a different one.

### Clean up on failure

If any step fails, return the error *and* delete the temp file, or failed runs
slowly fill the directory with `app.conf.tmp-…` junk. A `defer` with a success
flag handles every error path at once:

```go
ok := false
defer func() {
	if !ok {
		tmp.Close()           // harmless if already closed
		os.Remove(tmp.Name()) // throw away the half-made file
	}
}()
// ... each step: if err != nil { return err } ...
ok = true
return nil
```

### Permissions

`CreateTemp` always makes `0o600` files, and the rename keeps the temp file's
mode. So before the rename, `Chmod` the temp file:

- if the target already exists, use its current mode, so a `0o600` secrets
  file stays private: `info.Mode().Perm()` from `os.Stat(path)`;
- otherwise use the `perm` argument.

### How the tests check "replaced, not rewritten"

`os.Stat` returns an `fs.FileInfo`. `os.SameFile(a, b)` reports whether two
`FileInfo`s describe the same file on disk (on Linux, the same *inode*). A
plain `os.WriteFile` rewrites the old file in place, so before and after are
the same file. A rename swaps in a brand-new file, so they differ. The tests
stat the target before and after your write and require
`!os.SameFile(before, after)`.

## Your task

Your config tool needs a `WriteFileAtomic(path, data, perm)` with the same
signature as `os.WriteFile`, but safe. Add the imports you need (`os`,
`path/filepath`).

### 1. Write through a temp file and rename

Follow the recipe: create the temp file in `filepath.Dir(path)`, write `data`,
`Sync`, `Close`, then `os.Rename` it to `path`. Return the first error. After
a successful call, `path` holds exactly `data`, an existing file has been
replaced (not rewritten in place), and no temp file is left in the directory.

```go
WriteFileAtomic(dir+"/app.conf", []byte("port = 8080\n"), 0o644)
// dir holds only app.conf, containing "port = 8080\n"
```

### 2. Clean up when something fails

If any step fails, return a non-nil error and remove the temp file, so the
directory holds exactly what it held before. Two failures the tests use:

- `path` is an existing, non-empty directory, so the rename fails;
- the parent directory doesn't exist, so `CreateTemp` fails.

### 3. Set the permissions

Before the rename, `Chmod` the temp file. If `path` already exists, keep its
current permission bits, whatever `perm` says. If it doesn't exist, use `perm`
exactly.

```go
// secret.env exists with mode 0o600
WriteFileAtomic(dir+"/secret.env", data, 0o644) // still 0o600
// new.conf doesn't exist
WriteFileAtomic(dir+"/new.conf", data, 0o640)   // now 0o640
```

## Learn more

- [os.CreateTemp](https://pkg.go.dev/os#CreateTemp)
- [os.Rename](https://pkg.go.dev/os#Rename)
- [os.File.Sync](https://pkg.go.dev/os#File.Sync)
- [os.SameFile](https://pkg.go.dev/os#SameFile)
