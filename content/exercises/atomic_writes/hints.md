## Follow the diagram

Do the steps from the README in order, checking each error: `CreateTemp`,
`Write`, `Sync`, `Close`, `Rename`. Use `filepath.Dir(path)` as the temp
file's directory and `tmp.Name()` as the rename source.

## Task 1: the happy path

```go
tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
if err != nil {
	return err
}
if _, err := tmp.Write(data); err != nil {
	return err
}
// ... tmp.Sync(), tmp.Close(), the same way ...
return os.Rename(tmp.Name(), path)
```

## Task 2: one defer for every failure

Right after `CreateTemp` succeeds, set `ok := false` and defer a function that
closes and removes the temp file when `!ok`. Change the last line to check the
`Rename` error, then set `ok = true` and `return nil`.

## Task 3: pick the mode first

At the very top, before `CreateTemp`:

```go
mode := perm
if info, err := os.Stat(path); err == nil {
	mode = info.Mode().Perm()
}
```

Then call `tmp.Chmod(mode)` before `Sync`, checking its error.
