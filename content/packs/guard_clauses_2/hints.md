## Task 1: Fix the disk alert boundaries

The original said `< 80` means "ok". The guard should ask the opposite: "is it 80 or more?" With `>`, the value 80 itself slips through to "ok".

### Closer

The opposite of `<` is `>=`. Change both guards' operator; keep the numbers.

### Answer

```go
if usedPercent >= 95 {
	return "critical"
}
if usedPercent >= 80 {
	return "warning"
}
return "ok"
```

## Task 2: Fix the readiness guard

The original kept a pod when it was ready **and** had few restarts. So a pod must be skipped when either half fails, not only when both do.

### Closer

De Morgan: `!(a && b)` is `!a || !b`. The halves `!pod.Ready` and `pod.Restarts >= 3` are already right; only the operator between them changes.

### Answer

```go
if !pod.Ready || pod.Restarts >= 3 {
	continue
}
```

## Task 3: Make the down-host guard leave

The down-host guard adds the reason, then carries on into the disk check and the `append` to `patch`. A guard in a loop leaves with `continue`.

### Closer

Add `continue` as the last line inside `if !host.Up { ... }`, exactly like the disk guard below it.

### Answer

```go
if !host.Up {
	skipped = append(skipped, host.Name+": down")
	continue
}
```
