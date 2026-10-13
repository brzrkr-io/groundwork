## Task 1: Move the thresholds into disk.Status

The old `main` already has this logic. Check the biggest threshold first, so a value of 95 doesn't stop at the `warn` check.

### Closer

Use a `switch { ... }` with `case percent >= 90:` first, then `case percent >= 80:`, and `default:` for `"ok"`.

### Answer

```go
func Status(percent int) string {
	switch {
	case percent >= 90:
		return "full"
	case percent >= 80:
		return "warn"
	default:
		return "ok"
	}
}
```

## Task 2: Move the parsing into disk.Parse

The `strings.Cut` and `strconv.Atoi` lines from the old `main` move over almost unchanged. The difference is how a problem is reported: return an error instead of printing and exiting.

### Closer

Replace each `fmt.Fprintf(os.Stderr, ...)` plus `os.Exit(2)` with `return Reading{}, fmt.Errorf(...)`, without the `diskcheck: ` prefix and the newline. Use `%q` for the argument.

### Answer

```go
func Parse(arg string) (Reading, error) {
	host, value, found := strings.Cut(arg, "=")
	if !found || host == "" {
		return Reading{}, fmt.Errorf("bad reading %q: want HOST=PERCENT", arg)
	}
	percent, err := strconv.Atoi(value)
	if err != nil || percent < 0 || percent > 100 {
		return Reading{}, fmt.Errorf("bad percent in %q: want 0-100", arg)
	}
	return Reading{Host: host, Percent: percent}, nil
}
```

## Task 3: Write run for good input

Do it in two passes: first parse every argument into a list of readings, then print each one and work out the exit code. Write to the `stdout` parameter, not `os.Stdout`.

### Closer

`fmt.Fprintf(stdout, "%s %d%% %s\n", r.Host, r.Percent, status)` prints one line. Start with `code := 0` and set it to `1` when a status is `"full"`. Then make `main` just `os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))`.

### Answer

```go
readings := make([]disk.Reading, 0, len(args))
for _, arg := range args {
	r, err := disk.Parse(arg)
	if err != nil {
		// Task 4
	}
	readings = append(readings, r)
}

code := 0
for _, r := range readings {
	status := disk.Status(r.Percent)
	if status == "full" {
		code = 1
	}
	fmt.Fprintf(stdout, "%s %d%% %s\n", r.Host, r.Percent, status)
}
return code
```

```go
func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
```

## Task 4: Handle bad input in run

The no-arguments check at the top of `run` shows the shape: write to `stderr`, then return `2`. Do the same inside the parse loop when `disk.Parse` fails.

### Closer

Print the error with the `diskcheck: ` prefix and a newline, using `%v` for `err`. Because parsing happens before any printing, returning here means nothing reaches `stdout`.

### Answer

```go
r, err := disk.Parse(arg)
if err != nil {
	fmt.Fprintf(stderr, "diskcheck: %v\n", err)
	return 2
}
```
