# if and else

In *Booleans* you computed yes-or-no answers. Now you act on them: page
someone, cap a number, pick a message. Go makes decisions with `if` and
`else`.

**New here:** `if`, `else`, `else if`, early return, the `if` short statement.

## Learn

### if

An `if` statement runs a block of code only when its condition is `true`:

```go
if memoryMB > 4096 {
	return "too big"
}
```

The *condition* is any boolean expression from *Booleans*. The *block* is
everything between `{` and `}`.

> **Why is it written like this?** Go requires the braces, even around one
> line, and the `{` must be on the same line as the `if`. Go does **not**
> want parentheses around the condition: write `if x > 5 {`, not
> `if (x > 5) {` (`gofmt` removes them). Indenting the block with a tab is
> convention; `gofmt` does it for you.

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
is returned, because that branch comes first. **The order of your checks
matters.**

> **Why is it written like this?** `} else {` must sit on the same line as
> the closing `}`. Go requires that; putting `else` on a new line is a
> compile error.

### Early return

A `return` ends the function right there. So you often don't need `else` at
all: handle the special cases first and return early, then write the normal
case at the end.

```go
func TimeoutOrDefault(seconds int) int {
	if seconds <= 0 {
		return 30
	}
	return seconds
}
```

This *early return* style keeps the main path unindented and easy to read.
It's a strong Go convention, not a rule.

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

`free` is computed once and then checked. It exists in the condition and in
every `if`, `else if` and `else` block of this chain, but **not** after the
last `}`.

> **Why is it written like this?** The `;` between the statement and the
> condition is required. Whether you use a short statement at all is your
> choice; it's handy for a helper value you only need in the `if`.

### Integer division

With two `int`s, `/` drops the remainder: `250 / 3` is `83`, `7 / 2` is `3`.
Dividing by zero crashes the program.

## Your task

You are writing the decision logic for a small ops dashboard.

### 1. Classify disk usage

**Goal:** make `DiskAlert` return `"ok"`, `"warning"` or `"critical"` for
`usedPercent`.

| `usedPercent` | Return |
|---|---|
| below 80 | `"ok"` |
| 80 to 89 | `"warning"` |
| 90 and above | `"critical"` |

```go
DiskAlert(42)  // => "ok"
DiskAlert(79)  // => "ok"
DiskAlert(80)  // => "warning"
DiskAlert(89)  // => "warning"
DiskAlert(90)  // => "critical"
DiskAlert(100) // => "critical"
```

**Think it through:**
- Only the first matching branch runs. Which check must come first so 95
  isn't labelled `"warning"`?
- The starter already ends with `return "ok"`. Which values are left over for
  it?

### 2. Clamp the replica count

**Goal:** make `ClampReplicas` keep `requested` between 1 and 10, using early
returns.

- a request below 1 gives `1`,
- a request above 10 gives `10`,
- anything else (including 1 and 10) is returned unchanged.

```go
ClampReplicas(-3) // => 1
ClampReplicas(0)  // => 1
ClampReplicas(1)  // => 1
ClampReplicas(4)  // => 4
ClampReplicas(10) // => 10
ClampReplicas(25) // => 10
```

**Think it through:**
- There are two special cases. Each can be its own `if` that returns.
- The starter's `return requested` is the normal case. Where should it stay?

### 3. Give scaling advice

**Goal:** make `ScaleAdvice` return `"no replicas"`, `"scale up"`,
`"scale down"` or `"steady"` based on the requests per replica.

1. If `replicas` is below 1, return `"no replicas"` straight away, before any
   division.
2. Otherwise compute the requests per replica, `totalRequests` divided by
   `replicas` (integer division), in an `if` short statement.
3. Above 100 per replica: `"scale up"`.
4. Below 20 per replica: `"scale down"`.
5. Otherwise (including exactly 100 and exactly 20): `"steady"`.

```go
ScaleAdvice(1000, 0)  // => "no replicas"
ScaleAdvice(1000, -1) // => "no replicas"
ScaleAdvice(1000, 5)  // => "scale up" (200 per replica)
ScaleAdvice(202, 2)   // => "scale up" (101)
ScaleAdvice(500, 5)   // => "steady" (exactly 100)
ScaleAdvice(300, 5)   // => "steady" (60)
ScaleAdvice(100, 5)   // => "steady" (exactly 20)
ScaleAdvice(50, 5)    // => "scale down" (10)
ScaleAdvice(59, 3)    // => "scale down" (19, remainder dropped)
```

**Think it through:**
- Why must the `replicas` check come before the division?
- Look at the `free` example above: what name will you give the per-replica
  value, and which two parameters make it?

## Learn more

- [A Tour of Go: If](https://go.dev/tour/flowcontrol/5)
- [A Tour of Go: If with a short statement](https://go.dev/tour/flowcontrol/6)
- [Go by Example: If/Else](https://gobyexample.com/if-else)
- [Effective Go: If](https://go.dev/doc/effective_go#if)
