## Task 1: Split the READY column

READY is two numbers with a `/` between them. Split first, then convert
each half, and stop at the first thing that goes wrong.

### Closer

`strings.Cut(s, "/")` returns the part before, the part after, and whether
it found the `/`. `strconv.Atoi` converts each part. For the error, use
`fmt.Errorf` with `%q` so the value appears in quotes, and add the
`fmt`, `strconv` and `strings` imports.

### Answer

```go
before, after, found := strings.Cut(s, "/")
if !found {
	return 0, 0, fmt.Errorf("READY %q has no /", s)
}
ready, err = strconv.Atoi(before)
if err != nil {
	return 0, 0, fmt.Errorf("READY %q: %w", s, err)
}
total, err = strconv.Atoi(after)
if err != nil {
	return 0, 0, fmt.Errorf("READY %q: %w", s, err)
}
return ready, total, nil
```

## Task 2: Parse one row

Split the row into columns, check there are enough, and read positions 0
to 3. The restart count is only the first part of `6 (2m14s ago)`.

### Closer

`fields := strings.Fields(line)` gives the columns. Check
`len(fields) < 5` first. Then `ParseReady(fields[1])`, `fields[2]` for the
status and `strconv.Atoi(fields[3])` for the restarts.

### Answer

```go
fields := strings.Fields(line)
if len(fields) < 5 {
	return Pod{}, fmt.Errorf("row %q has %d columns, want at least 5", line, len(fields))
}
ready, total, err := ParseReady(fields[1])
if err != nil {
	return Pod{}, err
}
restarts, err := strconv.Atoi(fields[3])
if err != nil {
	return Pod{}, fmt.Errorf("RESTARTS in row %q: %w", line, err)
}
return Pod{Name: fields[0], Ready: ready, Total: total, Status: fields[2], Restarts: restarts}, nil
```

## Task 3: List the Pods that need attention

Split the output into lines, skip the first one and any blank ones, parse
the rest with `ParseLine`, and keep the names of Pods that aren't fine.
"Fine" is short to write; "needs attention" is "not fine".

### Closer

`strings.Split(output, "\n")` gives the lines; `lines[1:]` skips the
header. A small helper keeps the rule readable:

```go
func fine(p Pod) bool {
	if p.Status == "Completed" {
		return true
	}
	return p.Status == "Running" && ...
}
```

### Answer

```go
lines := strings.Split(output, "\n")
var names []string
for _, line := range lines[1:] {
	if strings.TrimSpace(line) == "" {
		continue
	}
	p, err := ParseLine(line)
	if err != nil {
		return nil, err
	}
	if !fine(p) {
		names = append(names, p.Name)
	}
}
return names, nil
```

with `fine` returning `p.Status == "Completed"`, or
`p.Status == "Running" && p.Ready == p.Total && p.Restarts < 3`.
