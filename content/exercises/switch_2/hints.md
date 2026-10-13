## Task 1: empty cases do nothing

`case "warn":` with nothing under it matches `"warn"` and then the switch
ends, so the function reaches the final `return -1`. Two cases have this
problem.

## Task 1: merge the cases

Replace the pair with one line: `case "warn", "warning":`. Do the same for
`"fatal"` and `"panic"`.

## Task 2: the shape

```go
switch {
case strings.HasPrefix(region, "us-"):
	return "North America"
// ...eu- and ap-...
default:
	return "other"
}
```
