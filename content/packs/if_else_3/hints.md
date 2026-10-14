## Task 1: Check the deploy window

Translate the README sentence straight into a boolean: "not production, or
the hour is 10 to 15". You can return it directly, as in *Booleans*.

### Closer

"Not production" is `env != "production"`. "10 to 15" is
`hour >= 10 && hour < 16`. Join the two parts with `||`, and wrap the hour
part in parentheses.

### Answer

```go
return env != "production" || (hour >= 10 && hour < 16)
```

## Task 2: Write the deploy gate

Write one `if` per rule, in the README's order, each with its own `return`.
Put `return "go"` at the very end. Add `import "fmt"` for the number in the
message.

### Closer

Check "exactly one" before the general case, so you can use the singular.
For more than one, put the number in with
`fmt.Sprintf("blocked: %d checks failing", failingChecks)`: the fixed text
stays inside the quotes and `%d` marks where the number goes. For the window,
call your task 1 function and flip it with `!`.

### Answer

```go
if !testsPassed {
	return "blocked: tests failing"
}
if failingChecks == 1 {
	return "blocked: 1 check failing"
}
if failingChecks > 1 {
	return fmt.Sprintf("blocked: %d checks failing", failingChecks)
}
if !InDeployWindow(env, hour) {
	return "blocked: outside deploy window"
}
return "go"
```
