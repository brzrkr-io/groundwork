## Task 1: don't throw the error away

Replace `_` with `err`, then add `if err != nil { return 0, err }`. Add a
second `if` for `n < 0` that returns `errors.New(...)`.

## Task 2: check, then trim

```go
k, v, found := strings.Cut(label, "=")
k = strings.TrimSpace(k)
```

Return `"", "", false` if `!found` or `k == ""`.

## Task 3: one suffix at a time

```go
if num, found := strings.CutSuffix(size, "Mi"); found {
	n, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	return n, nil
}
```

Do the same for `Gi`, multiplying by 1024 after checking the error.
Anything left over at the end of the function is an error.

## Almost the answer for task 3

The `Gi` branch and the final error:

```go
if num, found := strings.CutSuffix(size, "Gi"); found {
	n, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	return n * 1024, nil
}
return 0, fmt.Errorf("unknown unit in %q", size)
```
