# if and else: the deploy gate

Before a release goes out, a *deploy gate* checks that it is safe and says
either "go" or why not. You will write the gate for your team's pipeline.

## Learn

### Recap

As in *if and else*, checks run from top to bottom and the first `return`
ends the function. A chain of guard clauses (`if problem { return reason }`)
followed by the happy path at the end is a natural fit for a gate: each check
blocks for one reason, and the order of the checks decides which reason
the user sees first. You can build text with `fmt.Sprintf`, and you can call
your own functions from inside an `if` condition.

## Your task

### 1. Check the deploy window

Production deploys are only allowed from 10:00 up to, but not including,
16:00, so somebody is around to fix things. Other environments can deploy at
any hour. `hour` runs from 0 to 23.

Change `InDeployWindow` so it returns `true` when `env` is not
`"production"`, or when the hour is 10 to 15.

```go
InDeployWindow("staging", 3)     // => true
InDeployWindow("production", 10) // => true
InDeployWindow("production", 15) // => true
InDeployWindow("production", 16) // => false
InDeployWindow("production", 9)  // => false
```

### 2. Write the deploy gate

Change `DeployGate` to run these checks **in this order** and return the
result of the first one that blocks:

1. If `testsPassed` is `false`, return `"blocked: tests failing"`.
2. If `failingChecks` is above 0, return `"blocked: N checks failing"`, with
   `N` replaced by the number. Use the singular for exactly one:
   `"blocked: 1 check failing"`.
3. If `InDeployWindow(env, hour)` is `false`, return
   `"blocked: outside deploy window"`.
4. Otherwise, return `"go"`.

```go
DeployGate("production", 11, false, 2) // => "blocked: tests failing"
DeployGate("production", 11, true, 1)  // => "blocked: 1 check failing"
DeployGate("staging", 3, true, 4)      // => "blocked: 4 checks failing"
DeployGate("production", 20, true, 0)  // => "blocked: outside deploy window"
DeployGate("staging", 20, true, 0)     // => "go"
```

The `fmt` package is not imported yet. Add `import "fmt"` below the
`package` line when you need `fmt.Sprintf`.

## Learn more

- [Go by Example: If/Else](https://gobyexample.com/if-else)
- [Effective Go: If](https://go.dev/doc/effective_go#if)
- [pkg.go.dev: fmt.Sprintf](https://pkg.go.dev/fmt#Sprintf)
