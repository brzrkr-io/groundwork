## Task 1: one boolean expression

"Not production, or the hour is 10 to 15" translates straight into
`env != "production" || (hour >= 10 && hour < 16)`. You can return it
directly, as in *Booleans*.

## Task 2: one guard per rule

Write one `if` per rule, in the README's order, each with its own `return`.
Put `return "go"` at the very end.

## Task 2: the singular

Check for exactly one before the general case:

```go
if failingChecks == 1 {
	return "blocked: 1 check failing"
}
if failingChecks > 1 {
	return fmt.Sprintf("blocked: %d checks failing", failingChecks)
}
```

## Task 2: the window check

Call your function from task 1 inside the condition and flip it with `!`:

```go
if !InDeployWindow(env, hour) {
	return "blocked: outside deploy window"
}
```
