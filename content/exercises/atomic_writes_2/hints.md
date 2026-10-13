## Task 1: loop, replace, or append

Range over `lines` with an index. When `strings.HasPrefix(line, key+"=")`,
set `lines[i]` and return. If the loop ends without a match, return
`append(lines, key+"="+value)`.

## Task 2: replace the empty data with a real read

Swap `var data []byte` for a read that tolerates a missing file:

```go
data, err := os.ReadFile(path)
if err != nil && !errors.Is(err, fs.ErrNotExist) {
	return err
}
```

The two lines after it stay as they are.

## Task 3: a defer with a success flag

Right after `CreateTemp` succeeds:

```go
ok := false
defer func() {
	if !ok {
		tmp.Close()
		os.Remove(tmp.Name())
	}
}()
```

Then check the `Rename` error instead of returning it directly, and set
`ok = true` just before `return nil`.
