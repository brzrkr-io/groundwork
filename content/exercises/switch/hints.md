## Task 1: one case per row of the table

Each row of the table becomes a `case`. Write `case "dev", "test":` for the
shared row and `default:` for "anything else".

## Task 2: import and switch

Change the import line to:

```go
import "strings"
```

Then start with `switch strings.ToUpper(method) {` and list the uppercase
methods, comma-separated, in two cases.

## Task 3: strictest first

Since cases run top to bottom, `case ms < 100:` must come before
`case ms < 500:`:

```go
switch {
case ms < 100:
	return "fast"
case ms < 500:
	return "ok"
// ...
}
```
