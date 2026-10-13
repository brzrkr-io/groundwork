## Task 1: Write through a temp file and rename

Do the README's recipe steps in order, checking each error: create a temp file next to `path`, write, sync, close, then rename it over `path`.

### Closer

Create the temp file with `os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")`. Then call `tmp.Write(data)`, `tmp.Sync()`, `tmp.Close()`, and finally `os.Rename(tmp.Name(), path)`.

### Answer

```go
tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
if err != nil {
	return err
}
if _, err := tmp.Write(data); err != nil {
	return err
}
if err := tmp.Sync(); err != nil {
	return err
}
if err := tmp.Close(); err != nil {
	return err
}
return os.Rename(tmp.Name(), path)
```

## Task 2: Clean up when something fails

Every `return err` after `CreateTemp` leaves a temp file behind. Instead of cleaning up at each return, register the cleanup once and switch it off only when everything worked.

### Closer

Right after `CreateTemp` succeeds, set `ok := false` and `defer` a function that, when `!ok`, closes the temp file and calls `os.Remove(tmp.Name())`. Check the `Rename` error instead of returning it directly.

### Answer

```go
ok := false
defer func() {
	if !ok {
		tmp.Close()
		os.Remove(tmp.Name())
	}
}()

// ... Write, Sync, Close as before ...
if err := os.Rename(tmp.Name(), path); err != nil {
	return err
}
ok = true
return nil
```

## Task 3: Set the permissions

Decide the mode before you create the temp file: `perm` for a new file, or the existing file's bits if `path` is already there. Then apply it to the temp file before the rename.

### Closer

`os.Stat(path)` succeeds when the file exists; `info.Mode().Perm()` gives its permission bits. Apply the mode with `tmp.Chmod(mode)` before `Sync`.

### Answer

```go
mode := perm
if info, err := os.Stat(path); err == nil {
	mode = info.Mode().Perm()
}
// ... after CreateTemp and Write:
if err := tmp.Chmod(mode); err != nil {
	return err
}
```
