## Task 1: one check at a time

`fields := strings.Fields(line)` gives you a slice. Check `len(fields) != 3`
first, then convert `fields[1]` with `strconv.Atoi`, then check the range, then
`strings.HasPrefix(fields[2], "/")`. Each check returns early.

## Task 1: the messages

```go
return Check{}, fmt.Errorf("check %q: want 3 fields, got %d", line, len(fields))
```

## Task 2: loop with an index

```go
var checks []Check
for i, line := range lines {
	if strings.TrimSpace(line) == "" {
		continue
	}
	c, err := ParseCheck(line)
	if err != nil {
		return nil, fmt.Errorf("line %d: %v", i+1, err)
	}
	checks = append(checks, c)
}
return checks, nil
```
