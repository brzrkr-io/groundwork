# if and else: the deploy gate

Before a release goes out, a *deploy gate* checks that it is safe and says
either "go" or why not. You will write the gate for your team's pipeline.

**New here:** nothing; practice with `if`, `&&`, `||`, `!` and `fmt.Sprintf`.

## Learn

### Recap: guard clauses

As in *if and else*, checks run from top to bottom, and the first `return`
ends the function. A *guard clause* is an `if` that returns early when
something is wrong. The happy path goes last.

```go
func CanRestart(healthy bool, users int) string {
	if !healthy {
		return "no: unhealthy"
	}
	if users > 0 {
		return "no: users online"
	}
	return "yes"
}
```

The order of the guards decides which reason the user sees first.

> **Why is it written like this?** Each `if` needs `{ }` braces, even for
> one line (Go requires). Early returns instead of nested `else` blocks are
> convention. Which check goes first is the author's choice; here the README
> decides it for you.

### Combining conditions

| Operator | Meaning | Example |
|---|---|---|
| `&&` | and: both must be true | `cpu > 50 && cpu < 90` |
| `\|\|` | or: at least one is true | `zone == "a" \|\| zone == "b"` |
| `!` | not: flips true/false | `!ready` |

Parentheses group parts, like in maths: `a || (b && c)`.

### Calling your own function in a condition

A function that returns `bool` can go straight into an `if`:
`if !IsWeekend(day) { ... }`. You write its name *and* pass it values.

### Putting a number into text

`fmt.Sprintf` builds a string. Fixed text stays inside the quotes; `%d`
marks where an `int` goes; the value comes after the comma, outside the
quotes:

```go
fmt.Sprintf("%d pods restarting", restarts) // restarts=3 => "3 pods restarting"
```

> **Why is it written like this?** You must `import "fmt"` below the
> `package` line before using it (Go requires: no import, no `fmt`).

## Your task

Your pipeline needs a gate that blocks unsafe deploys.

### 1. Check the deploy window

**Goal:** `InDeployWindow(env, hour)` returns `true` when deploying is allowed.

Production may deploy only from 10:00 up to, but not including, 16:00 (so
hours 10 to 15). Every other environment may deploy at any hour. `hour`
runs from 0 to 23.

```go
InDeployWindow("staging", 3)     // => true
InDeployWindow("production", 10) // => true
InDeployWindow("production", 15) // => true
InDeployWindow("production", 16) // => false
InDeployWindow("production", 9)  // => false
```

**Think it through:**
- Say the rule as a sentence: "not production, *or* the hour is …". Which operator joins the two halves?
- "10 to 15" has a lower and an upper limit. Which operator says "both"?

### 2. Write the deploy gate

**Goal:** `DeployGate(env, hour, testsPassed, failingChecks)` returns `"go"`
or the first reason the deploy is blocked.

Run these checks **in this order**; return at the first one that blocks:

1. `testsPassed` is `false` → `"blocked: tests failing"`
2. `failingChecks` is exactly 1 → `"blocked: 1 check failing"` (singular)
3. `failingChecks` is above 1 → `"blocked: N checks failing"`, with `N` the number
4. `InDeployWindow(env, hour)` is `false` → `"blocked: outside deploy window"`
5. otherwise → `"go"`

```go
DeployGate("production", 11, false, 2) // => "blocked: tests failing"
DeployGate("production", 11, true, 1)  // => "blocked: 1 check failing"
DeployGate("staging", 3, true, 4)      // => "blocked: 4 checks failing"
DeployGate("production", 20, true, 2)  // => "blocked: 2 checks failing"
DeployGate("production", 20, true, 0)  // => "blocked: outside deploy window"
DeployGate("staging", 20, true, 0)     // => "go"
```

`fmt` is not imported yet; add `import "fmt"` when you need `fmt.Sprintf`.

**Think it through:**
- One `if` per rule, each with its own `return`. Where does `"go"` go?
- In message 3, which part is fixed text and which part is the parameter `failingChecks`?
- For rule 4, call your task 1 function with `env` and `hour`, and flip the result.

## Learn more

- [Go by Example: If/Else](https://gobyexample.com/if-else)
- [Effective Go: If](https://go.dev/doc/effective_go#if)
- [pkg.go.dev: fmt.Sprintf](https://pkg.go.dev/fmt#Sprintf)
