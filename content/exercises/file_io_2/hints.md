## Task 1: Stop wiping the log

`os.Create` truncates the file every time. Look at the flags table in the README: a log should be write-only, created on first use, and every write should land at the end.

### Closer

Call `os.OpenFile(name, flags, 0o644)` and combine the three flags with `|`: `os.O_APPEND|os.O_CREATE|...`. At the end, replace `f.Close()` and `return nil` with a single `return f.Close()`.

### Answer

```go
f, err := os.OpenFile(filepath.Join(dir, LogName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
if err != nil {
	return err
}
// ... Fprintln as before ...
return f.Close()
```

## Task 2: Create the log directory

Before opening the log, the directory has to exist, including every missing parent. One function in `os` creates the whole chain and does nothing if it already exists.

### Closer

Call `os.MkdirAll(dir, 0o755)` at the top of `AppendAudit` and return its error if it fails.

### Answer

```go
if err := os.MkdirAll(dir, 0o755); err != nil {
	return err
}
```

## Task 3: Read a log that doesn't exist yet

When `os.Open` fails because the file (or its directory) is missing, that just means "no entries yet". Tell that case apart from real errors.

### Closer

Check `errors.Is(err, fs.ErrNotExist)` before the `err != nil` check and return `[]string{}, nil` there. Add `"errors"` and `"io/fs"` to the imports.

### Answer

```go
f, err := os.Open(filepath.Join(dir, LogName))
if errors.Is(err, fs.ErrNotExist) {
	return []string{}, nil
}
if err != nil {
	return nil, err
}
```
