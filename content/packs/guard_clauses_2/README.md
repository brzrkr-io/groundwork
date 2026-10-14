# Guard clauses: fix the flipped conditions

A teammate flattened three nested functions into guard clauses, as in
*Guard clauses*. The code looks much better, but the alerting tests went red.
Turning a condition around is where refactors go wrong, so this is a
review: compare each function with the original below and fix the guard.

**New here:** nothing new; practice inverting conditions and reviewing a
refactor.

## Learn

### Recap: inverting a condition

A guard asks the opposite of the original check, and leaves.

| Original | Inverted |
|---|---|
| `a < b` | `a >= b` |
| `a <= b` | `a > b` |
| `a && b` | `!a \|\| !b` |
| `a \|\| b` | `!a && !b` |

Two traps catch almost everyone:

- **The boundary.** The opposite of `< 80` is `>= 80`, not `> 80`. The
  value 80 itself must land on exactly one side.
- **The `&&`.** Flipping `a && b` means flipping both halves **and** the
  `&&` into `||` (*De Morgan's laws*). "Ready and healthy" fails when the
  pod is not ready **or** not healthy, not only when both are wrong.

### Check with one value per branch

To review a flipped condition, pick values at the edge and run them through
both versions in your head. For `< 80`, try 79, 80 and 81. If the old and
new code disagree on any of them, the inversion is wrong.

### A guard must leave

A guard handles a case **and leaves**: `return` in a function, `continue`
(or `break`) in a loop. A guard that does its work but doesn't leave falls
through into the code meant for the normal case.

## Your task

Here is the original code before the refactor. It was correct; the tests
pass on it.

```go
// Before: DiskAlert
if usedPercent < 80 {
	return "ok"
} else if usedPercent < 95 {
	return "warning"
} else {
	return "critical"
}

// Before: the loop in Routable
if pod.Ready && pod.Restarts < 3 {
	names = append(names, pod.Name)
}

// Before: the loop in PatchPlan
if host.Up {
	if host.DiskUsedPercent < 90 {
		patch = append(patch, host.Name)
	} else {
		skipped = append(skipped, host.Name+": disk full")
	}
} else {
	skipped = append(skipped, host.Name+": down")
}
```

Fix the flattened versions in `exercise.go`. Keep them flat: one guard per
case, no `else`.

### 1. Fix the disk alert boundaries

**Goal:** make `DiskAlert` return `"critical"` from 95% up, `"warning"` from
80% up, and `"ok"` below 80%.

```go
DiskAlert(79) // => "ok"
DiskAlert(80) // => "warning"
DiskAlert(95) // => "critical"
```

**Think it through:**
- Run 80 and 95 through the original and the flattened version. Where do they disagree?
- What is the exact opposite of `usedPercent < 95`?

### 2. Fix the readiness guard

**Goal:** make `Routable` skip every pod that is not ready **or** has
restarted 3 times or more.

```go
// api-1 ready 0 restarts, api-2 not ready, api-3 ready 3 restarts,
// api-4 not ready 7 restarts, api-5 ready 2 restarts
Routable(pods) // => ["api-1", "api-5"]
```

**Think it through:**
- `api-2` is not ready but has 0 restarts. Does the current guard skip it?
- Each half was inverted correctly. Which part of the condition wasn't?

### 3. Make the down-host guard leave

**Goal:** make `PatchPlan` skip a down host with only the reason
`"<name>: down"`, and never patch it.

A host that is down is skipped as down, even if its disk is also full:

```go
// web-1 up 40%, web-2 down, web-3 up 92%, web-4 down 97%, web-5 up 89%
PatchPlan(hosts)
// => patch ["web-1", "web-5"]
//    skipped ["web-2: down", "web-3: disk full", "web-4: down"]
```

**Think it through:**
- Follow `web-2` through the loop. After its reason is added, which line runs next?
- How does the disk guard leave? Do the same.

## Learn more

- [Go spec: Logical operators](https://go.dev/ref/spec#Logical_operators)
- [De Morgan's laws (Wikipedia)](https://en.wikipedia.org/wiki/De_Morgan%27s_laws)
- [Go spec: Continue statements](https://go.dev/ref/spec#Continue_statements)
