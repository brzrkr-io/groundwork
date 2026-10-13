## Which flags does a log need?

Look at the flags table in the README. A log should be write-only, created on
first use, and every write should land at the end of the file. Combine those
three flags with `|`.

## Task 1: the OpenFile call

```go
f, err := os.OpenFile(filepath.Join(dir, LogName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
```

At the end, replace `f.Close()` and `return nil` with `return f.Close()`.

## Tasks 2 and 3: MkdirAll and ErrNotExist

At the top of `AppendAudit`:

```go
if err := os.MkdirAll(dir, 0o755); err != nil {
	return err
}
```

In `ReadAudit`, check `errors.Is(err, fs.ErrNotExist)` before returning the
error, and return `[]string{}, nil` in that case. Add `"errors"` and
`"io/fs"` to the imports.
