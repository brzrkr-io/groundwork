## Task 1: Fix ParseReplicas

The `_` throws away the error from `strconv.Atoi`. Give that error a name and
check it, then add a second check for negative numbers.

### Closer

Write `n, err := strconv.Atoi(s)` and then `if err != nil { return 0, err }`.
Add `if n < 0` that returns `0` and an error made with `errors.New(...)`. Add
`errors` to the imports.

### Answer

```go
n, err := strconv.Atoi(s)
if err != nil {
	return 0, err
}
if n < 0 {
	return 0, errors.New("replicas cannot be negative")
}
return n, nil
```

## Task 2: Finish ParseLabel

`strings.Cut` already split the label for you. Trim the spaces from the key
first, then decide whether the label is bad: no `=`, or an empty key.

### Closer

`k = strings.TrimSpace(k)`. Then `if !found || k == ""` return `"", "", false`.
Otherwise return the key, the trimmed value `strings.TrimSpace(v)`, and `true`.

### Answer

Below the starter's `strings.Cut` line, replace the TODO and the old `return`:

```go
k = strings.TrimSpace(k)
if !found || k == "" {
	return "", "", false
}
return k, strings.TrimSpace(v), true
```

## Task 3: Write ParseMemoryMiB

Try one ending at a time. `strings.CutSuffix` tells you whether the size ends
in `Mi`; if not, try `Gi`. If neither matches, that's an error.

### Closer

```go
if num, found := strings.CutSuffix(size, "Mi"); found {
	n, err := strconv.Atoi(num)
	// check err, then return n, nil
}
```

Do the same for `"Gi"`, but return `n * 1024`. At the very end, return `0` and
an error from `fmt.Errorf`. Add `fmt` to the imports.

### Answer

```go
if num, found := strings.CutSuffix(size, "Mi"); found {
	n, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	return n, nil
}
if num, found := strings.CutSuffix(size, "Gi"); found {
	n, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	return n * 1024, nil
}
return 0, fmt.Errorf("unknown unit in %q", size)
```
