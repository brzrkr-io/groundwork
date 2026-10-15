## Task 1: Flatten Verdict

Work from the outside in. Each `if` has a "bad" branch (its `else`) that returns a reason. Turn each one into an `if` that checks for the bad case and returns that reason.

### Closer

The outer `if c.Service != ""` becomes `if c.Service == "" { return "missing service name" }`. Then `if c.Frozen { ... }`, then `if !c.TestsPassed { ... }`, then `if c.Approvals < 1 { ... }`. The production check is already a guard; keep it, then `return "ok"`.

### Answer

```go
func Verdict(c Change) string {
	if c.Service == "" {
		return "missing service name"
	}
	if c.Frozen {
		return "change freeze in effect"
	}
	if !c.TestsPassed {
		return "tests failed"
	}
	if c.Approvals < 1 {
		return "needs an approval"
	}
	if c.Environment == "production" && c.Approvals < 2 {
		return "production needs 2 approvals"
	}
	return "ok"
}
```

## Task 2: Flatten the patch loop

Each nested `if` says "keep going only if...". Invert each one into "skip this host if..." and use `continue` to skip.

### Closer

`if !host.Up { continue }`, then the tag check without the `!`, then the disk check inverted: `< 90` becomes `>= 90`. The `append` goes last, directly in the loop body, after all three guards.

### Answer

```go
for _, host := range hosts {
	if !host.Up {
		continue
	}
	if slices.Contains(host.Tags, "do-not-patch") {
		continue
	}
	if host.DiskUsedPercent >= 90 {
		continue
	}
	targets = append(targets, host.Name)
}
```

## Task 3: Flatten ParseReplicas

Handle each error the moment you can detect it: missing `=`, wrong key, not a number, out of range. Whatever is left at the bottom is the success case.

### Closer

`if !found`, `if key != "replicas"`, `if err != nil`, then the range check inverted with De Morgan: `if n < 0 || n > 50`. Each returns `0` and its error; the last line is `return n, nil`.

### Answer

```go
key, value, found := strings.Cut(setting, "=")
if !found {
	return 0, fmt.Errorf(`missing "=" in %q`, setting)
}
if key != "replicas" {
	return 0, fmt.Errorf("unknown setting %q", key)
}
n, err := strconv.Atoi(value)
if err != nil {
	return 0, fmt.Errorf("replicas %q is not a number", value)
}
if n < 0 || n > 50 {
	return 0, fmt.Errorf("replicas must be between 0 and 50, got %d", n)
}
return n, nil
```
