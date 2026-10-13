## Task 1: methods on BaseCheck

Inside the `BaseCheck` methods, `b` is the embedded struct, so `b.Name` and
`b.IntervalSeconds` are right there. For `Interval`, use an `if` to handle the
`0` case first.

## Task 2: build host:port

`HTTPCheck.Target` is one line: `return h.URL`. For `TCPCheck`, the port is an
`int`, so format it with `fmt.Sprintf` and `%d` (add `import "fmt"`), or
convert it with `strconv.Itoa`.

## Task 3: loop through the interface

Inside `DueNow`, each `c` is a `Check`, so you can call `c.Interval()`,
`c.CheckName()` and `c.Target()` without knowing which type it is. Start with
`due := []string{}` and `append` to it.

## Almost the answer

```go
func DueNow(checks []Check, elapsed int) []string {
	due := []string{}
	for _, c := range checks {
		if elapsed%c.Interval() == 0 {
			due = append(due, c.CheckName()+" -> "+c.Target())
		}
	}
	return due
}
```
