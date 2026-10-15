## Task 1: Parse one line

Split the line into words with `strings.Fields`, then run the four checks from the README in order. Each check returns `Check{}` and an error as soon as it fails.

### Closer

Every message starts with `check %q: `, filled in with the whole `line`:

```go
fields := strings.Fields(line)
if len(fields) != 3 {
	return Check{}, fmt.Errorf("check %q: want 3 fields, got %d", line, len(fields))
}
```

Then convert `fields[1]` with `strconv.Atoi`, check `1`–`65535`, and test the path with `strings.HasPrefix(fields[2], "/")`.

### Answer

```go
fields := strings.Fields(line)
if len(fields) != 3 {
	return Check{}, fmt.Errorf("check %q: want 3 fields, got %d", line, len(fields))
}
port, err := strconv.Atoi(fields[1])
if err != nil {
	return Check{}, fmt.Errorf("check %q: port %q is not a number", line, fields[1])
}
if port < 1 || port > 65535 {
	return Check{}, fmt.Errorf("check %q: port %d is out of range 1-65535", line, port)
}
if !strings.HasPrefix(fields[2], "/") {
	return Check{}, fmt.Errorf("check %q: path %q must start with \"/\"", line, fields[2])
}
return Check{Name: fields[0], Port: port, Path: fields[2]}, nil
```

## Task 2: Parse the whole file

Loop over the lines with their index, skip the blank ones, and call `ParseCheck` on the rest. The index starts at 0, but line numbers start at 1.

### Closer

Use `for i, line := range lines`, skip with `if strings.TrimSpace(line) == "" { continue }`, and on error return `nil, fmt.Errorf("line %d: %v", i+1, err)`. Collect good checks with `append`.

### Answer

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

Blank lines still count in `i`, so the line numbers match the file.
