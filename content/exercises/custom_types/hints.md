## Task 1: compare with the constants

Inside `IsValid`, `e` is an `Environment`. Compare it with `==` against
`Dev`, `Staging` and `Production`, and join the three checks with `||`
("or").

```go
return e == Dev || e == Staging || e == Production
```

## Task 2: one severity at a time

A `switch` on `s` keeps the rules readable. `Critical` always returns
`true`. For `Error`, the answer depends on the environment, so return the
result of a comparison:

```go
switch s {
case Critical:
	return true
case Error:
	return env == Production
}
return false
```

## Task 3: lower-case first

Call `strings.ToLower(s)` once, then switch on the result. That way
`"WARNING"`, `"Warning"` and `"warning"` all become `"warning"`. Don't
forget `import "strings"` under the `package` line.

## Task 3: almost the answer

```go
switch strings.ToLower(s) {
case "debug":
	return Debug, true
case "info":
	return Info, true
// ... warning, error and critical the same way
default:
	return Debug, false
}
```
