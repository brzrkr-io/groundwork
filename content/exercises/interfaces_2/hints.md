## Compare line by line

Put the interface and the method next to each other. The interface says
`Check() bool`. Look at the name of `HTTPCheck`'s method, and at the result
type of `TCPCheck`'s `Check` method.

## Let the compiler point at it

Add these two lines near the top of `exercise.go`:

```go
var _ HealthChecker = HTTPCheck{}
var _ HealthChecker = TCPCheck{}
```

The build error names the missing or mismatched method for each type.

## Tasks 1 and 2: the fixes

`HTTPCheck`'s method is called `Checked`; rename it to `Check`. `TCPCheck`'s
`Check` returns a `string`; make it return a `bool`. `Open` already is the
answer:

```go
func (c TCPCheck) Check() bool {
	return c.Open
}
```

## Task 3: flip the condition

`FailingChecks` keeps a check when `c.Check()` is `true`. It should keep it
when the check is *not* healthy: `if !c.Check() { ... }`.
