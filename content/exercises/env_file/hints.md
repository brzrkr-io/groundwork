## Task 1: look at the first and last byte

A string can be indexed like a slice of bytes: `value[0]` is the first byte and
`value[len(value)-1]` is the last. Quote characters are single bytes, so you
can compare them with `'"'` and `'\''`. Check for "no quote at the start"
first; that case returns `value` unchanged.

## Task 2: strings.Cut does the splitting

```go
line, _ = strings.CutPrefix(strings.TrimSpace(line), "export ")
key, value, found := strings.Cut(line, "=")
if !found {
	return "", "", ErrMissingEquals
}
```

Then trim `key` before the `validKey.MatchString(key)` check, and finish
with `Unquote(strings.TrimSpace(value))`.

## Task 3: keep going after a bad line

Inside the scan loop, an error doesn't end the function. Collect it instead:

```go
key, value, err := ParseLine(line)
if err != nil {
	errs = append(errs, fmt.Errorf("line %d: %w", n, err))
	continue
}
vars[key] = value
```

After the loop, check `sc.Err()` first, then
`if len(errs) > 0 { return nil, errors.Join(errs...) }`.

## Task 4: sort the keys, not the map

```go
keys := make([]string, 0, len(vars))
for key := range vars {
	keys = append(keys, key)
}
slices.Sort(keys)
```

Then loop over `keys` and append `key + "=" + vars[key]` to the result.
