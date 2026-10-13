## Task 1: Failed health checks

Count the `false` values, then turn the count into a percentage. Handle an
empty slice first, so you never divide by zero.

### Closer

Return 0 when `len(checks) == 0`. Range with `for _, passed := range checks`
and do `failed++` when `!passed`. Multiply before you divide:
`failed * 100 / ...`.

### Answer

```go
if len(checks) == 0 {
	return 0
}
failed := 0
for _, passed := range checks {
	if !passed {
		failed++
	}
}
return failed * 100 / len(checks)
```

Multiplying by 100 first keeps the integer division from dropping to 0.

## Task 2: Deploy summary

Keep one counter per kind of result. Range over `results` and add one to the
right counter, then build the line.

### Closer

Use a `switch r { case "ok": ok++ ... }` inside the loop. Build the line with
`fmt.Sprintf("%d ok, %d failed, %d skipped", ...)`, and add `import "fmt"` at
the top.

### Answer

```go
ok, failed, skipped := 0, 0, 0
for _, r := range results {
	switch r {
	case "ok":
		ok++
	case "failed":
		failed++
	case "skipped":
		skipped++
	}
}
return fmt.Sprintf("%d ok, %d failed, %d skipped", ok, failed, skipped)
```

## Task 3: DNS labels

A name with no dots has 1 label, and every dot adds one more. An empty name is
the one special case.

### Closer

Return 0 for `""`. Start a counter at 1, then range over the string with
`for _, r := range name` and compare each rune with `'.'`.

### Answer

```go
if name == "" {
	return 0
}
labels := 1
for _, r := range name {
	if r == '.' {
		labels++
	}
}
return labels
```
