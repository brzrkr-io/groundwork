# if and else

In *Booleans* you computed yes-or-no answers. Now you act on them: page
someone, cap a number, pick a message. Go makes decisions with `if` and
`else`.

## Learn

### if

An `if` statement runs a block of code only when its condition is `true`:

```go
if usedPercent >= 90 {
	return "critical"
}
```

Two rules that surprise people coming from other languages:

- The condition needs **no parentheses**: write `if x > 5 {`, not `if (x > 5) {`.
- The **braces are required**, even around a single line, and the `{` must be
  on the same line as the `if`.

### else and else if

`else` runs when the condition is `false`. `else if` adds another condition
to try. Go checks the conditions from top to bottom and runs only the
**first** block whose condition is true:

```go
if latencyMs < 100 {
	return "fast"
} else if latencyMs < 500 {
	return "ok"
} else {
	return "slow"
}
```

For `latencyMs` of 50, both `< 100` and `< 500` are true, but only `"fast"`
is returned, because that branch comes first. The order of your checks matters.

### Early return

When a branch ends with `return`, the function stops right there. So you
often don't need `else` at all. Handle the special cases first and return
early, then write the normal case at the end:

```go
func TimeoutOrDefault(seconds int) int {
	if seconds <= 0 {
		return 30
	}
	return seconds
}
```

This *early return* style keeps the main path of a function unindented and
easy to read. Go programmers use it everywhere.

### A short statement before the condition

An `if` can start with a short statement, separated from the condition by a
semicolon:

```go
if free := total - used; free < 10 {
	return "low"
} else if free < 50 {
	return "medium"
} else {
	return "plenty"
}
```

`free` is computed once and then checked. It exists only inside this `if`
statement: you can use it in the condition and in every `if`, `else if` and
`else` block, but **not** after the closing `}` of the last block. That keeps
a helper variable from leaking into the rest of the function.

## Your task

You are writing the decision logic for a small ops dashboard.

### 1. Classify disk usage

Change `DiskAlert` to return a label for `usedPercent`:

| `usedPercent` | Return |
|---|---|
| below 80 | `"ok"` |
| 80 to 89 | `"warning"` |
| 90 and above | `"critical"` |

```go
DiskAlert(42) // => "ok"
DiskAlert(80) // => "warning"
DiskAlert(90) // => "critical"
```

### 2. Clamp the replica count

People request a number of replicas, but the cluster allows only 1 to 10.
Change `ClampReplicas` so that:

- a request below 1 gives `1`,
- a request above 10 gives `10`,
- anything else is returned unchanged.

Use early returns for the two special cases.

```go
ClampReplicas(0)  // => 1
ClampReplicas(4)  // => 4
ClampReplicas(25) // => 10
```

### 3. Give scaling advice

Change `ScaleAdvice` to look at the load per replica:

1. If `replicas` is below 1, return `"no replicas"` straight away. (This
   also stops you from dividing by zero, which would crash the program.)
2. Otherwise work out `per := totalRequests / replicas`. With two `int`s,
   `/` is integer division and drops the remainder: `250 / 3` is `83`.
   Compute `per` with a short statement in the `if`.
3. If `per` is above 100, return `"scale up"`.
4. If `per` is below 20, return `"scale down"`.
5. Otherwise return `"steady"`.

```go
ScaleAdvice(1000, 0) // => "no replicas"
ScaleAdvice(1000, 5) // => "scale up" (200 per replica)
ScaleAdvice(300, 5)  // => "steady" (60 per replica)
ScaleAdvice(50, 5)   // => "scale down" (10 per replica)
```

Exactly 100 and exactly 20 per replica are both `"steady"`.

## Learn more

- [A Tour of Go: If](https://go.dev/tour/flowcontrol/5)
- [A Tour of Go: If with a short statement](https://go.dev/tour/flowcontrol/6)
- [Go by Example: If/Else](https://gobyexample.com/if-else)
- [Effective Go: If](https://go.dev/doc/effective_go#if)
