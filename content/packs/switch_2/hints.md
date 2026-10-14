## Task 1: Fix the level numbers

Look at `case "warn":`. Nothing is written under it. In Go, that case matches
`"warn"`, does nothing, and ends the switch, so the function reaches
`return -1`. Two pairs of cases have this problem.

### Closer

Put both names in one case, separated by a comma, and delete the empty case:
`case "warn", "warning":`. Do the same for `"fatal"` and `"panic"`.

### Answer

```go
case "warn", "warning":
	return 2
case "error":
	return 3
case "fatal", "panic":
	return 4
```

## Task 2: Group regions by area

Replace the `if` with a tagless `switch {`. Each `case` asks
`strings.HasPrefix(region, ...)` with one prefix, and `default` handles the
rest.

### Closer

The prefixes are fixed text, so they go in quotes (`"us-"`); `region` is the
parameter, so it doesn't:

```go
switch {
case strings.HasPrefix(region, "us-"):
	return "North America"
// ...eu- and ap-...
default:
	return "other"
}
```

### Answer

```go
switch {
case strings.HasPrefix(region, "us-"):
	return "North America"
case strings.HasPrefix(region, "eu-"):
	return "Europe"
case strings.HasPrefix(region, "ap-"):
	return "Asia Pacific"
default:
	return "other"
}
```
