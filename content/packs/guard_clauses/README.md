# Guard clauses

Deeply nested code reads like a maze: to know when line 30 runs, you track
five open `if`s in your head, and each `else` sits far from the `if` it
belongs to. A *guard clause* handles a special case first and leaves
early, so the main path runs straight down the left edge. Go code is
written this way almost everywhere.

**New here:** guard clauses, `continue` as a guard, inverting a condition.

## Learn

### The arrow shape

Nested checks drift to the right, like an arrow:

```go
if user != "" {
	if hasAccess(user) {
		if !locked {
			restart(service)
			return nil
		} else {
			return errors.New("locked")
		}
	} else {
		return errors.New("no access")
	}
} else {
	return errors.New("no user")
}
```

The error for "no user" is 12 lines below the check that causes it.

### Turn each check around

Ask the opposite question, handle the bad case, and leave:

```go
if user == "" {
	return errors.New("no user")
}
if !hasAccess(user) {
	return errors.New("no access")
}
if locked {
	return errors.New("locked")
}
restart(service)
return nil
```

Same behaviour. Each problem sits right next to its check, and the last
lines are the *happy path*: what happens when everything is fine. No
`else` is needed, because a guard that returns means the code below it
only runs when the check passed.

> **Why is it written like this?** **Go requires** nothing here; both
> shapes compile. Returning early is strong **convention**: Go's own
> style guide asks for the normal path at the left edge with errors
> handled first ("indent error flow").

### Inverting a condition

To turn `if X { ...main path... }` into a guard, you write `if not X {
leave }`. Flip every piece:

| Original | Inverted |
|---|---|
| `a == b` | `a != b` |
| `a < b` | `a >= b` |
| `a >= b` | `a < b` |
| `ok` | `!ok` |
| `a && b` | `!a \|\| !b` |
| `a \|\| b` | `!a && !b` |

The last two rows are *De Morgan's laws*. "In range" is
`n >= 0 && n <= 50`; "out of range" is `n < 0 || n > 50`. Notice that
`>=` flips to `<`, not `<=`.

### Guards in loops: continue

Inside a loop, a guard skips to the next item with `continue`:

```go
for _, pod := range pods {
	if pod.Phase != "Running" {
		continue
	}
	if pod.Restarts > 5 {
		continue
	}
	healthy = append(healthy, pod.Name)
}
```

`continue` ends this pass of the loop and starts the next one. The loop
body reads like a checklist of "skip it if...".

### How the tests check the shape

`structure_test.go` reads `exercise.go` with Go's `go/ast` package and
measures how deep the blocks nest. An `if` directly in the function is
depth 1; an `if` inside a `for` loop is depth 2. It also rejects any
`else`. The behaviour checks run first, with every case in the tables, so
a flipped condition shows up before the shape does.

## Your task

The deploy tools have three functions shaped like arrows. Flatten them.
What each one returns must not change.

### 1. Flatten Verdict

**Goal:** rewrite `Verdict` as guard clauses: at most depth 1, and no
`else`.

`Verdict` checks the rules in this order and returns the first that fails:

| Rule | Returns |
|---|---|
| `Service` is empty | `"missing service name"` |
| `Frozen` is true | `"change freeze in effect"` |
| `TestsPassed` is false | `"tests failed"` |
| fewer than 1 approval | `"needs an approval"` |
| `Environment` is `"production"` with fewer than 2 approvals | `"production needs 2 approvals"` |
| none of the above | `"ok"` |

**Think it through:**
- The outermost `if` checks `c.Service != ""`. What is its inverted check, and what does it return?
- The order of the guards matters. What should a change with failed tests *and* no approvals return?

### 2. Flatten the patch loop

**Goal:** rewrite the loop in `PatchTargets` with `continue` guards: at most
depth 2 (the loop plus one `if`), and no `else`.

`PatchTargets` keeps hosts that are up, not tagged `"do-not-patch"`, and
less than 90% full, in their original order:

```go
// web-1 up 40%, web-2 down, db-1 tagged do-not-patch, web-3 at 90%, web-4 up 89%
PatchTargets(hosts) // => ["web-1", "web-4"]
```

**Think it through:**
- For each nested `if`, what is the opposite check that means "skip this host"?
- `host.DiskUsedPercent < 90` inverted is not `> 90`. What is it?

### 3. Flatten ParseReplicas

**Goal:** rewrite `ParseReplicas` so every error returns early: at most
depth 1, and no `else`.

The results stay the same:

| `setting` | Returns |
|---|---|
| `"replicas=3"` | `3, nil` (0 and 50 are allowed too) |
| `"replicas=51"` | error `replicas must be between 0 and 50, got 51` |
| `"replicas=three"` | error `replicas "three" is not a number` |
| `"memory=512Mi"` | error `unknown setting "memory"` |
| `"replicas3"` | error `missing "=" in "replicas3"` |

**Think it through:**
- Go code usually checks `err != nil` and returns. Which `if` here checks `err == nil` instead?
- How do you write "out of range" for `n >= 0 && n <= 50`? See the table in *Inverting a condition*.

## Learn more

- [Go Code Review Comments: Indent error flow](https://go.dev/wiki/CodeReviewComments#indent-error-flow)
- [Refactoring catalog: Replace Nested Conditional with Guard Clauses](https://refactoring.com/catalog/replaceNestedConditionalWithGuardClauses.html)
- [Go by Example: For (break and continue)](https://gobyexample.com/for)
