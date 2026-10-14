## Task 1: Write the shared methods

Inside the `BaseCheck` methods, `b` is the embedded struct, so `b.Name` and
`b.IntervalSeconds` are right there. `Interval` has a special case for `0`.

### Closer

`CheckName` is one line: `return b.Name`. In `Interval`, handle the zero case
first with `if b.IntervalSeconds == 0 { ... }`, then return the field.

### Answer

```go
func (b BaseCheck) CheckName() string {
	return b.Name
}

func (b BaseCheck) Interval() int {
	if b.IntervalSeconds == 0 {
		return DefaultInterval
	}
	return b.IntervalSeconds
}
```

## Task 2: Write each check's Target

`HTTPCheck` already stores the whole target in one field. For `TCPCheck`, the
port is an `int`, so it has to be turned into text before you can join it.

### Closer

`HTTPCheck.Target` is `return h.URL`. For `TCPCheck`, add `import "fmt"` and
use `fmt.Sprintf` with `"%s:%d"`.

### Answer

```go
func (h HTTPCheck) Target() string {
	return h.URL
}

func (c TCPCheck) Target() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
```

## Task 3: Find the checks that are due

Loop over the checks; each one is a `Check`, so you can call its three methods
without knowing its type. A check is due when the remainder of `elapsed`
divided by its interval is `0`.

### Closer

Start with `due := []string{}` so you return an empty slice, not `nil`. Inside
`for _, c := range checks`, test `elapsed%c.Interval() == 0` and `append` the
line `c.CheckName() + " -> " + c.Target()`.

### Answer

```go
due := []string{}
for _, c := range checks {
	if elapsed%c.Interval() == 0 {
		due = append(due, c.CheckName()+" -> "+c.Target())
	}
}
return due
```
