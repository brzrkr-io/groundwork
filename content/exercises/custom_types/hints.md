## Task 1: Validate an environment

`IsValid` should be `true` only when `e` equals one of the three constants. Compare `e` against each constant and combine the results.

### Closer

Use `==` for each comparison and join them with `||` ("or"): `e == Dev || ...`.

### Answer

```go
return e == Dev || e == Staging || e == Production
```

`Environment("prod")` matches none of the constants, so the whole expression is `false`.

## Task 2: Decide whether to page

Go through the table row by row. Only `Critical` and `Error` can ever page, and only `Error` cares about the environment.

### Closer

A `switch s` with a `case Critical:` and a `case Error:` keeps it readable. For `Error`, return the comparison itself: `return env == Production`.

### Answer

```go
switch s {
case Critical:
	return true
case Error:
	return env == Production
default:
	return false
}
```

Every other severity falls into `default` and never pages.

## Task 3: Parse a severity

Lower-case the text first, so `"WARNING"`, `"Warning"` and `"warning"` all look the same. Then match the five names. Remember to add `import "strings"` under the `package` line.

### Closer

Switch on `strings.ToLower(s)` and return a pair from each case:

```go
switch strings.ToLower(s) {
case "debug":
	return Debug, true
// ... info, warning, error, critical
}
```

### Answer

```go
switch strings.ToLower(s) {
case "debug":
	return Debug, true
case "info":
	return Info, true
case "warning":
	return Warning, true
case "error":
	return Error, true
case "critical":
	return Critical, true
default:
	return Debug, false
}
```

`default` catches unknown names like `"fatal"` and returns the zero value with `false`.
