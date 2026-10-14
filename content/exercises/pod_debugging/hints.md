## Task 1: Decode an exit code

Exit codes above 128 are "killed by a signal", and the signal's number is
what's left after taking 128 away. Anything 128 or below wasn't a signal.

### Closer

One `if code > 128` with two returns: `code - 128, true` inside it, and
`0, false` after it.

### Answer

```go
if code > 128 {
	return code - 128, true
}
return 0, false
```

## Task 2: Read the last termination

Split the output into lines and look for the `Last State:` line. From
there, keep reading while the lines are indented deeper than that line,
and pick out two fields. If you never find the line, return
`ErrNoTermination`.

### Closer

For each line: `key, value, _ := strings.Cut(strings.TrimSpace(line), ":")`.
The start is `key == "Last State"` with the trimmed value `"Terminated"`.
Count indentation with a helper:

```go
func indent(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}
```

Inside the block, `switch key` on `"Reason"` and `"Exit Code"`, then
`strconv.Atoi` the exit code. Add the `fmt`, `strconv` and `strings`
imports.

### Answer

```go
lines := strings.Split(describe, "\n")
for i, line := range lines {
	key, value, _ := strings.Cut(strings.TrimSpace(line), ":")
	if key != "Last State" || strings.TrimSpace(value) != "Terminated" {
		continue
	}
	var t Termination
	exitCode := ""
	depth := indent(line)
	for _, field := range lines[i+1:] {
		if indent(field) <= depth {
			break
		}
		key, value, _ := strings.Cut(strings.TrimSpace(field), ":")
		switch key {
		case "Reason":
			t.Reason = strings.TrimSpace(value)
		case "Exit Code":
			exitCode = strings.TrimSpace(value)
		}
	}
	code, err := strconv.Atoi(exitCode)
	if err != nil {
		return Termination{}, fmt.Errorf("last state exit code %q: %w", exitCode, err)
	}
	t.ExitCode = code
	return t, nil
}
return Termination{}, ErrNoTermination
```

## Task 3: Name the cause

The table is already a program: one check per row, top to bottom, return
on the first match. Call `LastTermination` once at the top and remember
whether it found a block.

### Closer

A `switch` with no value runs the first `case` whose condition is true:

```go
last, err := LastTermination(describe)
restarted := err == nil
switch {
case strings.Contains(describe, "ErrImagePull") || strings.Contains(describe, "ImagePullBackOff"):
	return CauseImage
case restarted && last.Reason == "OOMKilled":
	...
}
return CauseNone
```

### Answer

```go
has := func(s string) bool { return strings.Contains(describe, s) }
last, err := LastTermination(describe)
restarted := err == nil

switch {
case has("ErrImagePull") || has("ImagePullBackOff"):
	return CauseImage
case has("FailedScheduling"):
	return CauseUnschedulable
case has("CreateContainerConfigError"):
	return CauseConfig
case restarted && last.Reason == "OOMKilled":
	return CauseMemory
case has("failed liveness probe"):
	return CauseLiveness
case restarted && last.ExitCode != 0:
	return CauseCrash
case has("Readiness probe failed"):
	return CauseReadiness
}
return CauseNone
```
