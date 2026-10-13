## Task 1: Parse a container image reference

Split the reference at the `:` with `strings.Cut`. Then check the cases in
order: an empty name, no `:` at all, and a `:` with nothing after it.

### Closer

`name, tag, found := strings.Cut(ref, ":")`. If `name == ""`, return an error.
If `!found`, return `name`, the fixed text `"latest"`, and `nil`. If `tag == ""`,
return an error. Add `errors`, `fmt` and `strings` to the imports.

### Answer

```go
name, tag, found := strings.Cut(ref, ":")
if name == "" {
	return "", "", errors.New("image name is empty")
}
if !found {
	return name, "latest", nil
}
if tag == "" {
	return "", "", fmt.Errorf("empty tag in %q", ref)
}
return name, tag, nil
```

## Task 2: Parse a cache TTL

First find the unit at the end, and remember how many seconds it is worth (1,
60 or 3600). Then turn the rest into a number and check it is above 0.

### Closer

Use `strings.CutSuffix` once per unit in an `if` / `else if` chain, saving the
rest and the multiplier:

```go
if rest, found := strings.CutSuffix(ttl, "s"); found {
	num, multiplier = rest, 1
} else if ...
```

Then `strconv.Atoi(num)`, check the error, reject `n < 1`, and return
`n * multiplier, nil`. Add `strconv` to the imports.

### Answer

```go
multiplier := 0
num := ""
if rest, found := strings.CutSuffix(ttl, "s"); found {
	num, multiplier = rest, 1
} else if rest, found := strings.CutSuffix(ttl, "m"); found {
	num, multiplier = rest, 60
} else if rest, found := strings.CutSuffix(ttl, "h"); found {
	num, multiplier = rest, 3600
} else {
	return 0, fmt.Errorf("unknown unit in %q", ttl)
}
n, err := strconv.Atoi(num)
if err != nil {
	return 0, err
}
if n < 1 {
	return 0, fmt.Errorf("TTL must be above 0, got %q", ttl)
}
return n * multiplier, nil
```
