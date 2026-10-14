## Task 1: Update the lines in memory

Look through `lines` for the one that belongs to `key`. If you find it, replace it in place; only append when no line matched.

### Closer

Range with an index: `for i, line := range lines`. A line belongs to `key` when `strings.HasPrefix(line, key+"=")`. Set `lines[i]` and return early; after the loop, append.

### Answer

```go
for i, line := range lines {
	if strings.HasPrefix(line, key+"=") {
		lines[i] = key + "=" + value
		return lines
	}
}
return append(lines, key+"="+value)
```

## Task 2: Read, modify, write

`SetKey` starts from empty `data`. Replace that with a real read of the file, treating a missing file as empty.

### Closer

Use `os.ReadFile(path)`. Only return the error when it is not a "not exist" error: `err != nil && !errors.Is(err, fs.ErrNotExist)`. Add `"errors"` and `"io/fs"` to the imports.

### Answer

```go
data, err := os.ReadFile(path)
if err != nil && !errors.Is(err, fs.ErrNotExist) {
	return err
}
lines := setLine(splitLines(string(data)), key, value)
```

## Task 3: Fix the leaking helper

Every `return err` after `CreateTemp` in `writeFileAtomic` leaves the temp file behind. Register the cleanup once, and switch it off only after the rename succeeds.

### Closer

Right after `CreateTemp`, set `ok := false` and `defer` a function that closes and removes `tmp.Name()` when `!ok`. Then check the `Rename` error instead of returning it directly, and set `ok = true` before `return nil`.

### Answer

```go
ok := false
defer func() {
	if !ok {
		tmp.Close()
		os.Remove(tmp.Name())
	}
}()

// ... Write, Chmod, Sync, Close as before ...
if err := os.Rename(tmp.Name(), path); err != nil {
	return err
}
ok = true
return nil
```
